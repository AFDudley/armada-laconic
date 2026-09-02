package nitro

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/statechannels/go-nitro/node/query"
	nitrotypes "github.com/statechannels/go-nitro/types"

	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/server/relay"
)

func OpenChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-channel <counterparty> <amount>",
		Short: "Open a ledger channel with the given counterparty",
		Args:  cobra.ExactArgs(2),
		Long:  "Creates a ledger channel with the specified counterparty and funding amount.",
		RunE: func(cmd *cobra.Command, args []string) error {
			counterparty, err := nitrotypes.ParseParticipant(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse counterparty: %w", err)
			}

			clientCtx, s, done, err := setupNitroCommand(cmd)
			if err != nil {
				return err
			}
			defer done()

			assetAddr, _ := cmd.Flags().GetString("asset")
			funds, err := resolveTokens(s, args[1], assetAddr)
			if err != nil {
				return err
			}

			m, err := s.OpenLedgerChannel(funds, counterparty)
			if err != nil {
				return err
			}

			if err = waitForObjective(
				s.node.FailedObjectives(),
				m.Objective,
				s.node.LedgerUpdatedChan(m.Objective.ChannelID()),
				func(info query.LedgerChannelInfo) bool { return info.Status == query.Open },
			); err != nil {
				return err
			}

			return submitNitroTx(*clientCtx, cmd, *m, false)
		},
	}

	cmd.Flags().String("asset", "", "Ethereum token address (if specified, amount is parsed as decimal without denomination)")
	AddFlags(cmd.Flags())
	flags.AddTxFlagsToCmd(cmd)

	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func CloseChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-channel <channel-id>",
		Short: "Close a ledger channel",
		Args:  cobra.ExactArgs(1),
		Long:  "Close a ledger channel. Use --challenge flag for dispute resolution when cooperation fails.",
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse channel ID: %w", err)
			}

			isChallenge, err := cmd.Flags().GetBool("challenge")
			if err != nil {
				return err
			}

			clientCtx, s, done, err := setupNitroCommand(cmd)
			if err != nil {
				return err
			}
			defer done()

			m, err := s.CloseLedgerChannel(channelId, isChallenge)
			if err != nil {
				return err
			}

			if err = waitForObjective(
				s.node.FailedObjectives(),
				m.Objective,
				s.node.LedgerUpdatedChan(channelId),
				func(info query.LedgerChannelInfo) bool { return info.Status == query.Complete },
			); err != nil {
				return err
			}

			return submitNitroTx(*clientCtx, cmd, *m, false)
		},
	}

	cmd.Flags().Bool("challenge", false, "Close via challenge (dispute resolution) instead of cooperative closure")
	AddFlags(cmd.Flags())
	flags.AddTxFlagsToCmd(cmd)

	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func OpenPaymentChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-payment-channel <counterparty> <amount> [intermediaries...]",
		Short: "Open a virtual payment channel",
		Args:  cobra.MinimumNArgs(2),
		Long:  "Creates a virtual payment channel with the specified counterparty and optional intermediaries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			counterparty := args[0]

			challengeDuration, err := cmd.Flags().GetUint32("challenge-duration")
			if err != nil {
				return err
			}

			var intermediaries []string
			if len(args) > 2 {
				intermediaries = args[2:]
			}

			clientCtx, s, done, err := setupNitroCommand(cmd)
			if err != nil {
				return err
			}
			defer done()

			assetAddr, _ := cmd.Flags().GetString("asset")
			funds, err := resolveTokens(s, args[1], assetAddr)
			if err != nil {
				return err
			}

			m, err := s.CreatePaymentChannel(intermediaries, counterparty, challengeDuration, funds)
			if err != nil {
				return err
			}

			if err = waitForObjective(
				s.node.FailedObjectives(),
				m.Objective,
				s.node.PaymentChannelUpdatedChan(m.Objective.ChannelID()),
				func(info query.PaymentChannelInfo) bool { return info.Status == query.Open },
			); err != nil {
				return err
			}

			return submitNitroTx(*clientCtx, cmd, *m, false)
		},
	}

	cmd.Flags().String("asset", "", "Ethereum token address (if specified, amount is parsed as decimal without denomination)")
	cmd.Flags().Uint32("challenge-duration", 0, "Challenge duration for the channel")
	AddFlags(cmd.Flags())
	flags.AddTxFlagsToCmd(cmd)

	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func ClosePaymentChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-payment-channel <channel-id>",
		Short: "Close a virtual payment channel",
		Args:  cobra.ExactArgs(1),
		Long:  "Close a virtual payment channel with the given channel ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse channel ID: %w", err)
			}

			clientCtx, s, done, err := setupNitroCommand(cmd)
			if err != nil {
				return err
			}
			defer done()

			m, err := s.ClosePaymentChannel(channelId)
			if err != nil {
				return err
			}

			if err = waitForObjective(
				s.node.FailedObjectives(),
				m.Objective,
				s.node.PaymentChannelUpdatedChan(channelId),
				func(info query.PaymentChannelInfo) bool { return info.Status == query.Complete },
			); err != nil {
				return err
			}

			return submitNitroTx(*clientCtx, cmd, *m, false)
		},
	}

	AddFlags(cmd.Flags())
	flags.AddTxFlagsToCmd(cmd)

	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func PayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pay <channel-id> <amount>",
		Short: "Send a payment through a payment channel",
		Args:  cobra.ExactArgs(2),
		Long:  "Creates and sends a payment voucher through the specified payment channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse channel ID: %w", err)
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse amount: %w", err)
			}
			if len(coins) != 1 {
				return fmt.Errorf("exactly one coin denomination required")
			}

			clientCtx, s, done, err := setupNitroCommand(cmd)
			if err != nil {
				return err
			}
			defer done()

			info, err := s.node.GetPaymentChannel(channelId)
			if err != nil {
				return err
			}
			paidSoFar := info.Balance.PaidSoFar.ToInt()
			expectedAmount := new(big.Int).Add(paidSoFar, coins[0].Amount.BigInt())
			updateChan := s.node.PaymentChannelUpdatedChan(channelId)

			m, err := s.Pay(channelId, coins[0])
			if err != nil {
				return fmt.Errorf("failed to send payment: %w", err)
			}

			// TODO: unordered tx for payments: this could break if tx is sent before opening the
			// channel; think through wrt. state integration
			if err = submitNitroTx(*clientCtx, cmd, *m, true); err != nil {
				return err
			}

			for info := range updateChan {
				if expectedAmount.Cmp(info.Balance.PaidSoFar.ToInt()) <= 0 {
					break
				}
			}
			delayAfterUpdate()
			return nil
		},
	}

	AddFlags(cmd.Flags())
	flags.AddTxFlagsToCmd(cmd)

	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}

func CreateVoucherCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-voucher <channel-id> <amount>",
		Short: "Create a payment voucher for a payment channel",
		Args:  cobra.ExactArgs(2),
		Long:  "Creates a signed payment voucher that can be used to redeem payments from the specified channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId, err := nitrotypes.ParseChannelID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse channel ID: %w", err)
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse amount: %w", err)
			}

			if len(coins) != 1 {
				return fmt.Errorf("exactly one coin denomination required")
			}
			amount := coins[0].Amount.BigInt()

			s, done, err := setupNitroQueryCommand(cmd, false)
			if err != nil {
				return err
			}
			defer done()

			voucher, err := s.node.CreateVoucher(channelId, amount)
			if err != nil {
				return fmt.Errorf("failed to create voucher: %w", err)
			}

			output, err := json.MarshalIndent(voucher, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal voucher: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	server.SetRequiredComponents(cmd, &Server{}, &relay.Server{})
	return cmd
}
