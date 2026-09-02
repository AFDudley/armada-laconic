package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"git.vdb.to/cerc-io/laconicd/server/nitro"
	types "git.vdb.to/cerc-io/laconicd/x/nitro/types"
)

// NewQueryCmd returns a root CLI handler for all x/nitro query commands.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		nitro.GetChannelCmd(),
		nitro.ListChannelsCmd(),
		nitro.ListPaymentChannelsCmd(),
		nitro.IdentityCmd(),
	)

	return cmd
}
