package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

// GetTxCmd returns transaction commands for this module.
func GetTxCmd() *cobra.Command {
	registryTxCmd := &cobra.Command{
		Use:                        registrytypes.ModuleName,
		Short:                      "registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	registryTxCmd.AddCommand(
		GetCmdSetRecord(),
	)

	return registryTxCmd
}

// GetCmdSetRecord is the CLI command for creating/updating a record.
func GetCmdSetRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [payload-file-path] [bond-id]",
		Short: "Set record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			payloadType, err := GetPayloadFromFile(args[0])
			if err != nil {
				return err
			}

			payload := payloadType.ToPayload()

			msg := registrytypes.NewMsgSetRecord(payload, args[1], clientCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetPayloadFromFile loads payload object from YAML file.
func GetPayloadFromFile(filePath string) (*registrytypes.ReadablePayload, error) {
	var payload registrytypes.ReadablePayload

	data, err := os.ReadFile(filePath) // #nosec G304
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

// GetCmdReserveAuthority is the CLI command for reserving a name.
func GetCmdReserveAuthority() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reserve-authority [name]",
		Short: "Reserve authority name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := registrytypes.NewMsgReserveAuthority(args[0], clientCtx.GetFromAddress(), ownerAddress)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().String("owner", "", "Owner address, if creating a sub-authority.")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSetAuthorityBond is the CLI command for associating a bond with an authority.
func GetCmdSetAuthorityBond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authority-bond [name] [bond-id]",
		Short: "Associate authority with bond.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Reserver name with owner address .
Example:
$ %s tx %s authority-bond [name] [bond-id]
`,
				version.AppName, registrytypes.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := registrytypes.NewMsgSetAuthorityBond(args[0], args[1], clientCtx.GetFromAddress())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSetName is the CLI command for mapping a name to a CID.
func GetCmdSetName() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-name [crn] [cid]",
		Short: "Set CRN to CID mapping.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set name with crn and cid.
Example:
$ %s tx %s set-name [crn] [cid]
`,
				version.AppName, registrytypes.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := registrytypes.NewMsgSetName(args[0], args[1], clientCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
