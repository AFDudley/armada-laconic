package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"git.vdb.to/cerc-io/laconicd/server/nitro"
	"git.vdb.to/cerc-io/laconicd/x/nitro/types"
)

// NewTxCmd returns a root CLI command handler for all x/nitro transaction commands.
func NewTxCmd() *cobra.Command {
	nitroTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Nitro transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nitroTxCmd.AddCommand(
		nitro.OpenChannelCmd(),
		nitro.CloseChannelCmd(),
		nitro.OpenPaymentChannelCmd(),
		nitro.ClosePaymentChannelCmd(),
		nitro.PayCmd(),
	)

	return nitroTxCmd
}
