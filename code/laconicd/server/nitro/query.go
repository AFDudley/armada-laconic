package nitro

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	nitrotypes "github.com/statechannels/go-nitro/types"

	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/server/relay"
)

func GetChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-channel <channel-id>",
		Short: "Get information about a channel",
		Args:  cobra.ExactArgs(1),
		Long:  "Retrieves detailed information about a specific ledger or payment channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse channel ID: %w", err)
			}

			ledger, _ := cmd.Flags().GetBool("ledger")
			payment, _ := cmd.Flags().GetBool("payment")
			local, _ := cmd.Flags().GetBool("local")

			if !ledger && !payment {
				ledger = true
			}
			if ledger && payment {
				return fmt.Errorf("cannot specify both --ledger and --payment flags")
			}

			s, done, err := setupNitroQueryCommand(cmd, !local)
			if err != nil {
				return err
			}
			defer done()

			var output []byte
			if ledger {
				info, err := s.node.GetLedgerChannel(channelId)
				if err != nil {
					return fmt.Errorf("failed to get ledger channel: %w", err)
				}
				output, err = json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal response: %w", err)
				}
			} else {
				info, err := s.node.GetPaymentChannel(channelId)
				if err != nil {
					return fmt.Errorf("failed to get payment channel: %w", err)
				}
				output, err = json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal response: %w", err)
				}
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().Bool("ledger", false, "Get ledger channel information")
	cmd.Flags().Bool("payment", false, "Get payment channel information")
	cmd.Flags().Bool("local", false, "Query local data only (no P2P connection)")
	flags.AddQueryFlagsToCmd(cmd)
	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func ListChannelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-channels",
		Short: "List all ledger channels",
		Args:  cobra.NoArgs,
		Long:  "Retrieves information about all ledger channels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			local, _ := cmd.Flags().GetBool("local")

			s, done, err := setupNitroQueryCommand(cmd, !local)
			if err != nil {
				return err
			}
			defer done()

			channels, err := s.node.GetAllLedgerChannels()
			if err != nil {
				return fmt.Errorf("failed to get all ledger channels: %w", err)
			}

			output, err := json.MarshalIndent(channels, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal response: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().Bool("local", false, "Query local data only (no P2P connection)")
	flags.AddQueryFlagsToCmd(cmd)
	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func ListPaymentChannelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-payment-channels <ledger-channel-id>",
		Short: "List all payment channels for a ledger channel",
		Args:  cobra.ExactArgs(1),
		Long:  "Retrieves all active payment channels associated with a specific ledger channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ledgerChannelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse ledger channel ID: %w", err)
			}

			local, _ := cmd.Flags().GetBool("local")

			s, done, err := setupNitroQueryCommand(cmd, !local)
			if err != nil {
				return err
			}
			defer done()

			channels, err := s.node.GetPaymentChannelsByLedger(ledgerChannelId)
			if err != nil {
				return fmt.Errorf("failed to get payment channels: %w", err)
			}

			output, err := json.MarshalIndent(channels, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal response: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().Bool("local", false, "Query local data only (no P2P connection)")
	flags.AddQueryFlagsToCmd(cmd)
	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func IdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Get the node identity",
		Args:  cobra.NoArgs,
		Long:  "Retrieves the node's identity. Use --participant for principal ID (default) or --network for network address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			participant, _ := cmd.Flags().GetBool("participant")
			network, _ := cmd.Flags().GetBool("network")

			if !participant && !network {
				participant = true
			}
			if participant && network {
				return fmt.Errorf("cannot specify both --participant and --network flags")
			}

			s, done, err := setupNitroQueryCommand(cmd, false)
			if err != nil {
				return err
			}
			defer done()

			if participant {
				identity := s.node.PrincipalID()
				fmt.Println(identity.String())
			} else {
				identity := s.node.NetworkID()
				fmt.Println(identity.String())
			}
			return nil
		},
	}

	cmd.Flags().Bool("participant", false, "Get the principal ID (default)")
	cmd.Flags().Bool("network", false, "Get the network address")
	flags.AddQueryFlagsToCmd(cmd)
	server.SetRequiredComponents(cmd, &Server{})
	return cmd
}
