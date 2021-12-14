package starportcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/clispinner"
	"github.com/tendermint/starport/starport/pkg/placeholder"
)

func NewScaffoldWasm() *cobra.Command {
	c := &cobra.Command{
		Use:   "wasm",
		Short: "Import the wasm module to your app",
		Long:  "Add support for WebAssembly smart contracts to your blockchain",
		Args:  cobra.NoArgs,
		RunE:  scaffoldWasmHandler,
	}

	flagSetPath(c)

	return c
}

func scaffoldWasmHandler(cmd *cobra.Command, args []string) error {
	appPath := flagGetPath(cmd)

	s := clispinner.New().SetText("Scaffolding...")
	defer s.Stop()

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.ImportModule(placeholder.New(), "wasm")
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 Imported wasm.\n\n")

	return nil
}
