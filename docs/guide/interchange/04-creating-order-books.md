---
order: 4
---
# Implement the Order Books

In this chapter you will implement the logic to create order books.

In Cosmos SDK the state is stored in a key-value store. Each order book will be stored under a unique key composed of four values: port ID, channel ID, source denom and target denom. For example, an order book for `mcx` and `vcx` could be stored under `ibcdex-channel-4-mcx-vcx`. Define a function that returns an order book store key.

```go
// x/ibcdex/types/keys.go
import "fmt"

//...
func OrderBookIndex( portID string, channelID string, sourceDenom string, targetDenom string, ) string {
  return fmt.Sprintf("%s-%s-%s-%s", portID, channelID, sourceDenom, targetDenom, )
}
```

`send-create-pair` is used to create order books. This command creates and broadcasts a transaction with a message of type `SendCreatePair`. The message gets routed to the `ibcdex` module, processed by the message handler in `x/ibcdex/handler.go` and finally a `SendCreatePair` keeper method is called.

You need `send-create-pair` to do the following:

* When processing `SendCreatePair` message on the source chain
  * Check that an order book with the given pair of denoms does not yet exist
  * Transmit an IBC packet with information about port, channel, source and target denoms
* Upon receiving the packet on the target chain
  * Check that an order book with the given pair of denoms does not yet exist on the target chain
  * Create a new order book for buy orders
  * Transmit an IBC acknowledgement back to the source chain
* Upon receiving the acknowledgement on the source chain
  * Create a new order book for sell orders

## Message Handling in SendCreatePair

`SendCreatePair` function was created during the IBC packet scaffolding. Currently, it creates an IBC packet, populates it with source and target denoms and transmits this packet over IBC. Add the logic to check for an existing order book for a particular pair of denoms.

```go
// x/ibcdex/keeper/msg_server_create_pair.go
import (
  "errors"
  //...
)

func (k msgServer) SendCreatePair(goCtx context.Context, msg *types.MsgSendCreatePair) (*types.MsgSendCreatePairResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Get an order book index
	pairIndex := types.OrderBookIndex(msg.Port, msg.ChannelID, msg.SourceDenom, msg.TargetDenom)
	// If an order book is found, return an error
	_, found := k.GetSellOrderBook(ctx, pairIndex)
	if found {
		return &types.MsgSendCreatePairResponse{}, errors.New("the pair already exist")
	}

	// Construct the packet
	var packet types.CreatePairPacketData

	packet.SourceDenom = msg.SourceDenom
	packet.TargetDenom = msg.TargetDenom

	// Transmit the packet
	err := k.TransmitCreatePairPacket(
		ctx,
		packet,
		msg.Port,
		msg.ChannelID,
		clienttypes.ZeroHeight(),
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendCreatePairResponse{}, nil
}
```

## Lifecycle of an IBC Packet

During a successful transmission, an IBC packet goes through 4 stages:

1. Message processing before packet transmission (on the source cahin)
2. Reception of a packet (on the target chain)
3. Acknowledgment of a packet (on the source chain)
4. Timeout of a packet (on the source chain)

In the following section you'll be implementing packet reception logic in the `OnRecvCreatePairPacket` function and packet acknowledgement logic in the `OnAcknowledgementCreatePairPacket` function. Timeout function will be left empty.

## Receiving an IBC packet

The protocol buffer definition defines the data that an order book has. Add the `OrderBook` and `Order` messages to the `order.proto` file.
First you will need to add the proto buffer files. This builds the according go code files that you can then modify for the purpose of your app.

Create a new `order.proto` file in the `proto/ibcdex` directory and add the content.

```proto
// proto/ibcdex/order.proto
syntax = "proto3";
package cosmonaut.interchange.ibcdex;

option go_package = "github.com/cosmonaut/interchange/x/ibcdex/types";

message OrderBook {
  int32 idCount = 1;
  repeated Order orders = 2;
}

message Order {
  int32 id = 1;
  string creator = 2;
  int32 amount = 3;
  int32 price = 4;
}
```

Modify the `buy_order_book.proto` file to have the fields for creating a buy order on the order book.

```proto
// proto/ibcdex/buy_order_book.proto
import "ibcdex/order.proto";

message BuyOrderBook {
  // ...
  OrderBook book = 5;
}
```

Modify the `sell_order_book.proto` file to add the order book into the buy order book. The proto definition for the `SellOrderBook` should look like follows:

```proto
// proto/ibcdex/sell_order_book.proto
// ...
import "ibcdex/order.proto";

message SellOrderBook {
  // ...
  OrderBook book = 6;
}
```

Now build the proto files to go with the command:

```bash
starport generate proto-go
```

Start enhancing the functions for the IBC packets.

Create a new file `x/ibcdex/types/order_book.go`.
Add the new order book function to the corresponing Go file.

```go
// x/ibcdex/types/order_book.go
package types

func NewOrderBook() OrderBook {
	return OrderBook{
		IdCount: 0,
	}
}
```

Define `NewBuyOrderBook` in a new file `x/ibcdex/types/buy_order_book.go` creates a new buy order book.

```go
// x/ibcdex/types/buy_order_book.go
package types

func NewBuyOrderBook(AmountDenom string, PriceDenom string) BuyOrderBook {
	book := NewOrderBook()
	return BuyOrderBook{
		AmountDenom: AmountDenom,
		PriceDenom: PriceDenom,
		Book: &book,
	}
}
```

On the target chain when an IBC packet is recieved, the module should check whether a book already exists, if not, create a new buy order book for specified denoms.

```go
// x/ibcdex/keeper/create_pair.go
func (k Keeper) OnRecvCreatePairPacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreatePairPacketData) (packetAck types.CreatePairPacketAck, err error) {
  // ...
  // Get an order book index
  pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.SourceDenom, data.TargetDenom)
  // If an order book is found, return an error
  _, found := k.GetBuyOrderBook(ctx, pairIndex)
  if found {
    return packetAck, errors.New("the pair already exist")
  }
  // Create a new buy order book for source and target denoms
  book := types.NewBuyOrderBook(data.SourceDenom, data.TargetDenom)
  // Assign order book index
  book.Index = pairIndex
  // Save the order book to the store
  k.SetBuyOrderBook(ctx, book)
  return packetAck, nil
}
```

## Receiving an IBC Acknowledgement

On the source chain when an IBC acknowledgement is recieved, the module should check whether a book already exists, if not, create a new sell order book for specified denoms.

Create a new file `x/ibcdex/types/sell_order_book.go`.
Insert the `NewSellOrderBook` function which creates a new sell order book.

```go
// x/ibcdex/types/sell_order_book.go
package types

func NewSellOrderBook(AmountDenom string, PriceDenom string) SellOrderBook {
	book := NewOrderBook()
	return SellOrderBook{
		AmountDenom: AmountDenom,
		PriceDenom: PriceDenom,
		Book: &book,
	}
}
```

Modify the Acknowledgement function in the `create_pair.go` file.

```go
// x/ibcdex/keeper/create_pair.go
func (k Keeper) OnAcknowledgementCreatePairPacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreatePairPacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.CreatePairPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}
		// Set the sell order book
		pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.SourceDenom, data.TargetDenom)
		book := types.NewSellOrderBook(data.SourceDenom, data.TargetDenom)
		book.Index = pairIndex
		k.SetSellOrderBook(ctx, book)
		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}
```

In this chapter we implemented the logic behind `send-create-pair` command that upon recieving of an IBC packet on the target chain creates a buy order book and upon recieving of an IBC acknowledgement on the source chain creates a sell order book.

### Implement the `appendOrder` Function to Add Orders to the Order Book

```go
// x/ibcdex/types/order_book.go
package types

import (
	"errors"
  "sort"
)

const (
	MaxAmount = int32(100000)
	MaxPrice  = int32(100000)
)

type Ordering int

const (
	Increasing Ordering = iota
	Decreasing
)

var (
	ErrMaxAmount     = errors.New("max amount reached")
	ErrMaxPrice      = errors.New("max price reached")
	ErrZeroAmount    = errors.New("amount is zero")
	ErrZeroPrice     = errors.New("price is zero")
	ErrOrderNotFound = errors.New("order not found")
)
```

`AppendOrder` initializes and appends a new order to an order book from the order information.

```go
// x/ibcdex/types/order_book.go
func (book *OrderBook) appendOrder(creator string, amount int32, price int32, ordering Ordering) (int32, error) {
	if err := checkAmountAndPrice(amount, price); err != nil {
		return 0, err
	}
	// Initialize the order
	var order Order
	order.Id = book.GetNextOrderID()
	order.Creator = creator
	order.Amount = amount
	order.Price = price
	// Increment ID tracker
	book.IncrementNextOrderID()
	// Insert the order
	book.insertOrder(order, ordering)
	return order.Id, nil
}
```

#### Implement the checkAmountAndPrice For an Order

`checkAmountAndPrice` checks correct amount or price.

```go
// x/ibcdex/types/order_book.go
func checkAmountAndPrice(amount int32, price int32) error {
	if amount == int32(0) {
		return ErrZeroAmount
	}
	if amount > MaxAmount {
		return ErrMaxAmount
	}
	if price == int32(0) {
		return ErrZeroPrice
	}
	if price > MaxPrice {
		return ErrMaxPrice
	}
	return nil
}
```

#### Implement the GetNextOrderID Function

`GetNextOrderID` gets the ID of the next order to append

```go
// x/ibcdex/types/order_book.go
func (book OrderBook) GetNextOrderID() int32 {
	return book.IdCount
}
```

#### Implement the IncrementNextOrderID Function

`IncrementNextOrderID` updates the ID count for orders

```go
// x/ibcdex/types/order_book.go
func (book *OrderBook) IncrementNextOrderID() {
	// Even numbers to have different ID than buy orders
	book.IdCount++
}
```

#### Implement the insertOrder Function

`insertOrder` inserts the order in the book with the provided order

```go
// x/ibcdex/types/order_book.go
func (book *OrderBook) insertOrder(order Order, ordering Ordering) {
	if len(book.Orders) > 0 {
		var i int
		// get the index of the new order depending on the provided ordering
		if ordering == Increasing {
			i = sort.Search(len(book.Orders), func(i int) bool { return book.Orders[i].Price > order.Price })
		} else {
			i = sort.Search(len(book.Orders), func(i int) bool { return book.Orders[i].Price < order.Price })
		}
		// insert order
		orders := append(book.Orders, &order)
		copy(orders[i+1:], orders[i:])
		orders[i] = &order
		book.Orders = orders
	} else {
		book.Orders = append(book.Orders, &order)
	}
}
```

This completes the order book setup. In the next chapter, you will learn how to deal with vouchers, minting and burning vouchers as well as locking and unlocking native blockchain token in your app.
