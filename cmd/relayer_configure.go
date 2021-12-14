package starportcmd

import (
	"fmt"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/cliquiz"
	"github.com/tendermint/starport/starport/pkg/clispinner"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/relayer"
)

const (
	flagAdvanced            = "advanced"
	flagSourceAccount       = "source-account"
	flagTargetAccount       = "target-account"
	flagSourceRPC           = "source-rpc"
	flagTargetRPC           = "target-rpc"
	flagSourceFaucet        = "source-faucet"
	flagTargetFaucet        = "target-faucet"
	flagSourcePort          = "source-port"
	flagSourceVersion       = "source-version"
	flagTargetPort          = "target-port"
	flagTargetVersion       = "target-version"
	flagSourceGasPrice      = "source-gasprice"
	flagTargetGasPrice      = "target-gasprice"
	flagSourceGasLimit      = "source-gaslimit"
	flagTargetGasLimit      = "target-gaslimit"
	flagSourceAddressPrefix = "source-prefix"
	flagTargetAddressPrefix = "target-prefix"
	flagOrdered             = "ordered"

	relayerSource = "source"
	relayerTarget = "target"

	defaultSourceRPCAddress = "http://localhost:26657"
	defaultTargetRPCAddress = "https://rpc.cosmos.network:443"

	defautSourceGasPrice      = "0.00025stake"
	defautTargetGasPrice      = "0.025uatom"
	defautSourceGasLimit      = 300000
	defautTargetGasLimit      = 300000
	defautSourceAddressPrefix = "cosmos"
	defautTargetAddressPrefix = "cosmos"
)

// NewRelayerConfigure returns a new relayer configure command.
// faucet addresses are optional and connect command will try to guess the address
// when not provided. even if auto retrieving coins fails, connect command will complete with success.
func NewRelayerConfigure() *cobra.Command {
	c := &cobra.Command{
		Use:     "configure",
		Short:   "Configure source and target chains for relaying",
		Aliases: []string{"conf"},
		RunE:    relayerConfigureHandler,
	}
	c.Flags().BoolP(flagAdvanced, "a", false, "Advanced configuration options for custom IBC modules")
	c.Flags().String(flagSourceRPC, "", "RPC address of the source chain")
	c.Flags().String(flagTargetRPC, "", "RPC address of the target chain")
	c.Flags().String(flagSourceFaucet, "", "Faucet address of the source chain")
	c.Flags().String(flagTargetFaucet, "", "Faucet address of the target chain")
	c.Flags().String(flagSourcePort, "", "IBC port ID on the source chain")
	c.Flags().String(flagSourceVersion, "", "Module version on the source chain")
	c.Flags().String(flagTargetPort, "", "IBC port ID on the target chain")
	c.Flags().String(flagTargetVersion, "", "Module version on the target chain")
	c.Flags().String(flagSourceGasPrice, "", "Gas price used for transactions on source chain")
	c.Flags().String(flagTargetGasPrice, "", "Gas price used for transactions on target chain")
	c.Flags().Int64(flagSourceGasLimit, 0, "Gas limit used for transactions on source chain")
	c.Flags().Int64(flagTargetGasLimit, 0, "Gas limit used for transactions on target chain")
	c.Flags().String(flagSourceAddressPrefix, "", "Address prefix of the source chain")
	c.Flags().String(flagTargetAddressPrefix, "", "Address prefix of the target chain")
	c.Flags().String(flagSourceAccount, "", "Source Account")
	c.Flags().String(flagTargetAccount, "", "Target Account")
	c.Flags().Bool(flagOrdered, false, "Set the channel as ordered")
	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func relayerConfigureHandler(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		err = handleRelayerAccountErr(err)
	}()

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	if err := ca.EnsureDefaultAccount(); err != nil {
		return err
	}

	s := clispinner.New().Stop()
	defer s.Stop()

	printSection("Setting up chains")

	// basic configuration
	var (
		sourceAccount       string
		targetAccount       string
		sourceRPCAddress    string
		targetRPCAddress    string
		sourceFaucetAddress string
		targetFaucetAddress string
		sourceGasPrice      string
		targetGasPrice      string
		sourceGasLimit      int64
		targetGasLimit      int64
		sourceAddressPrefix string
		targetAddressPrefix string
	)

	// advanced configuration for the channel
	var (
		sourcePort    string
		sourceVersion string
		targetPort    string
		targetVersion string
	)

	// questions
	var (
		questionSourceAccount = cliquiz.NewQuestion(
			"Source Account",
			&sourceAccount,
			cliquiz.DefaultAnswer(cosmosaccount.DefaultAccount),
			cliquiz.Required(),
		)
		questionTargetAccount = cliquiz.NewQuestion(
			"Target Account",
			&targetAccount,
			cliquiz.DefaultAnswer(cosmosaccount.DefaultAccount),
			cliquiz.Required(),
		)
		questionSourceRPCAddress = cliquiz.NewQuestion(
			"Source RPC",
			&sourceRPCAddress,
			cliquiz.DefaultAnswer(defaultSourceRPCAddress),
			cliquiz.Required(),
		)
		questionSourceFaucet = cliquiz.NewQuestion(
			"Source Faucet",
			&sourceFaucetAddress,
		)
		questionTargetRPCAddress = cliquiz.NewQuestion(
			"Target RPC",
			&targetRPCAddress,
			cliquiz.DefaultAnswer(defaultTargetRPCAddress),
			cliquiz.Required(),
		)
		questionTargetFaucet = cliquiz.NewQuestion(
			"Target Faucet",
			&targetFaucetAddress,
		)
		questionSourcePort = cliquiz.NewQuestion(
			"Source Port",
			&sourcePort,
			cliquiz.DefaultAnswer(relayer.TransferPort),
			cliquiz.Required(),
		)
		questionSourceVersion = cliquiz.NewQuestion(
			"Source Version",
			&sourceVersion,
			cliquiz.DefaultAnswer(relayer.TransferVersion),
			cliquiz.Required(),
		)
		questionTargetPort = cliquiz.NewQuestion(
			"Target Port",
			&targetPort,
			cliquiz.DefaultAnswer(relayer.TransferPort),
			cliquiz.Required(),
		)
		questionTargetVersion = cliquiz.NewQuestion(
			"Target Version",
			&targetVersion,
			cliquiz.DefaultAnswer(relayer.TransferVersion),
			cliquiz.Required(),
		)
		questionSourceGasPrice = cliquiz.NewQuestion(
			"Source Gas Price",
			&sourceGasPrice,
			cliquiz.DefaultAnswer(defautSourceGasPrice),
			cliquiz.Required(),
		)
		questionTargetGasPrice = cliquiz.NewQuestion(
			"Target Gas Price",
			&targetGasPrice,
			cliquiz.DefaultAnswer(defautTargetGasPrice),
			cliquiz.Required(),
		)
		questionSourceGasLimit = cliquiz.NewQuestion(
			"Source Gas Limit",
			&sourceGasLimit,
			cliquiz.DefaultAnswer(defautSourceGasLimit),
			cliquiz.Required(),
		)
		questionTargetGasLimit = cliquiz.NewQuestion(
			"Target Gas Limit",
			&targetGasLimit,
			cliquiz.DefaultAnswer(defautTargetGasLimit),
			cliquiz.Required(),
		)
		questionSourceAddressPrefix = cliquiz.NewQuestion(
			"Source Address Prefix",
			&sourceAddressPrefix,
			cliquiz.DefaultAnswer(defautSourceAddressPrefix),
			cliquiz.Required(),
		)
		questionTargetAddressPrefix = cliquiz.NewQuestion(
			"Target Address Prefix",
			&targetAddressPrefix,
			cliquiz.DefaultAnswer(defautTargetAddressPrefix),
			cliquiz.Required(),
		)
	)

	// Get flags
	advanced, err := cmd.Flags().GetBool(flagAdvanced)
	if err != nil {
		return err
	}
	sourceAccount, err = cmd.Flags().GetString(flagSourceAccount)
	if err != nil {
		return err
	}
	targetAccount, err = cmd.Flags().GetString(flagTargetAccount)
	if err != nil {
		return err
	}
	sourceRPCAddress, err = cmd.Flags().GetString(flagSourceRPC)
	if err != nil {
		return err
	}
	sourceFaucetAddress, err = cmd.Flags().GetString(flagSourceFaucet)
	if err != nil {
		return err
	}
	targetRPCAddress, err = cmd.Flags().GetString(flagTargetRPC)
	if err != nil {
		return err
	}
	targetFaucetAddress, err = cmd.Flags().GetString(flagTargetFaucet)
	if err != nil {
		return err
	}
	sourcePort, err = cmd.Flags().GetString(flagSourcePort)
	if err != nil {
		return err
	}
	sourceVersion, err = cmd.Flags().GetString(flagSourceVersion)
	if err != nil {
		return err
	}
	targetPort, err = cmd.Flags().GetString(flagTargetPort)
	if err != nil {
		return err
	}
	targetVersion, err = cmd.Flags().GetString(flagTargetVersion)
	if err != nil {
		return err
	}
	sourceGasPrice, err = cmd.Flags().GetString(flagSourceGasPrice)
	if err != nil {
		return err
	}
	targetGasPrice, err = cmd.Flags().GetString(flagTargetGasPrice)
	if err != nil {
		return err
	}
	sourceGasLimit, err = cmd.Flags().GetInt64(flagSourceGasLimit)
	if err != nil {
		return err
	}
	targetGasLimit, err = cmd.Flags().GetInt64(flagTargetGasLimit)
	if err != nil {
		return err
	}
	sourceAddressPrefix, err = cmd.Flags().GetString(flagSourceAddressPrefix)
	if err != nil {
		return err
	}
	targetAddressPrefix, err = cmd.Flags().GetString(flagTargetAddressPrefix)
	if err != nil {
		return err
	}
	ordered, err := cmd.Flags().GetBool(flagOrdered)
	if err != nil {
		return err
	}

	var questions []cliquiz.Question

	// get information from prompt if flag not provided
	if sourceAccount == "" {
		questions = append(questions, questionSourceAccount)
	}
	if targetAccount == "" {
		questions = append(questions, questionTargetAccount)
	}
	if sourceRPCAddress == "" {
		questions = append(questions, questionSourceRPCAddress)
	}
	if sourceFaucetAddress == "" {
		questions = append(questions, questionSourceFaucet)
	}
	if targetRPCAddress == "" {
		questions = append(questions, questionTargetRPCAddress)
	}
	if targetFaucetAddress == "" {
		questions = append(questions, questionTargetFaucet)
	}
	if sourceGasPrice == "" {
		questions = append(questions, questionSourceGasPrice)
	}
	if targetGasPrice == "" {
		questions = append(questions, questionTargetGasPrice)
	}
	if sourceGasLimit == 0 {
		questions = append(questions, questionSourceGasLimit)
	}
	if targetGasLimit == 0 {
		questions = append(questions, questionTargetGasLimit)
	}
	if sourceAddressPrefix == "" {
		questions = append(questions, questionSourceAddressPrefix)
	}
	if targetAddressPrefix == "" {
		questions = append(questions, questionTargetAddressPrefix)
	}
	// advanced information
	if advanced {
		if sourcePort == "" {
			questions = append(questions, questionSourcePort)
		}
		if sourceVersion == "" {
			questions = append(questions, questionSourceVersion)
		}
		if targetPort == "" {
			questions = append(questions, questionTargetPort)
		}
		if targetVersion == "" {
			questions = append(questions, questionTargetVersion)
		}
	}

	if len(questions) > 0 {
		if err := cliquiz.Ask(questions...); err != nil {
			return err
		}
	}

	r := relayer.New(ca)

	fmt.Println()
	s.SetText("Fetching chain info...")

	// initialize the chains
	sourceChain, err := initChain(
		cmd,
		r,
		s,
		relayerSource,
		sourceAccount,
		sourceRPCAddress,
		sourceFaucetAddress,
		sourceGasPrice,
		sourceGasLimit,
		sourceAddressPrefix,
	)
	if err != nil {
		return err
	}

	targetChain, err := initChain(
		cmd,
		r,
		s,
		relayerTarget,
		targetAccount,
		targetRPCAddress,
		targetFaucetAddress,
		targetGasPrice,
		targetGasLimit,
		targetAddressPrefix,
	)
	if err != nil {
		return err
	}

	s.SetText("Configuring...").Start()

	// sets advanced channel options
	var channelOptions []relayer.ChannelOption
	if advanced {
		channelOptions = append(channelOptions,
			relayer.SourcePort(sourcePort),
			relayer.SourceVersion(sourceVersion),
			relayer.TargetPort(targetPort),
			relayer.TargetVersion(targetVersion),
		)

		if ordered {
			channelOptions = append(channelOptions, relayer.Ordered())
		}
	}

	// create the connection configuration
	id, err := sourceChain.Connect(cmd.Context(), targetChain, channelOptions...)
	if err != nil {
		return err
	}

	s.Stop()

	fmt.Printf("⛓  Configured chains: %s\n\n", color.Green.Sprint(id))

	return nil
}

// initChain initializes chain information for the relayer connection
func initChain(
	cmd *cobra.Command,
	r relayer.Relayer,
	s *clispinner.Spinner,
	name,
	accountName,
	rpcAddr,
	faucetAddr,
	gasPrice string,
	gasLimit int64,
	addressPrefix string,
) (*relayer.Chain, error) {
	defer s.Stop()
	s.SetText("Initializing chain...").Start()

	c, account, err := r.NewChain(
		cmd.Context(),
		accountName,
		rpcAddr,
		relayer.WithFaucet(faucetAddr),
		relayer.WithGasPrice(gasPrice),
		relayer.WithGasLimit(gasLimit),
		relayer.WithAddressPrefix(addressPrefix),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot resolve %s", name)
	}

	s.Stop()

	accountAddr := account.Address(addressPrefix)

	fmt.Printf("🔐  Account on %q is %s(%s)\n \n", name, accountName, accountAddr)
	s.
		SetCharset(spinner.CharSets[9]).
		SetColor("white").
		SetPrefix(" |·").
		SetText(color.Yellow.Sprintf("trying to receive tokens from a faucet...")).
		Start()

	coins, err := c.TryRetrieve(cmd.Context())
	s.Stop()

	fmt.Print(" |· ")
	if err != nil {
		fmt.Println(color.Yellow.Sprintf(err.Error()))
	} else {
		fmt.Println(color.Green.Sprintf("received coins from a faucet"))
	}

	balance := coins.String()
	if balance == "" {
		balance = "-"
	}
	fmt.Printf(" |· (balance: %s)\n\n", balance)

	return c, nil
}

func printSection(title string) {
	fmt.Printf("---------------------------------------------\n%s\n---------------------------------------------\n\n", title)
}
