package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
)

var FlagSaveBidsTo = "save-to"

// GetTxCmd returns transaction commands for this module.
func GetTxCmd() *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:                        auctiontypes.ModuleName,
		Short:                      "Auction transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	auctionTxCmd.AddCommand(
		GetCmdCommitBid(),
		GetCmdRevealBid(),
	)

	return auctionTxCmd
}

// GetCmdCommitBid is the CLI command for committing a bid.
func GetCmdCommitBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-bid [auction-id] [bid-amount]",
		Short: "Commit sealed bid",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Take chain id passed by user
			chainId, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainId == "" {
				// Take from config if not provided
				chainId = clientCtx.ChainID
			}
			if chainId == "" {
				return fmt.Errorf("--chain-id required")
			}

			bidAmount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			mnemonic, err := utils.GenerateMnemonic()
			if err != nil {
				return err
			}

			auctionId := args[0]
			fromAddr := clientCtx.GetFromAddress().String()
			reveal := map[string]interface{}{
				"chainId":       chainId,
				"auctionId":     auctionId,
				"bidderAddress": fromAddr,
				"bidAmount":     bidAmount.String(),
				"noise":         mnemonic,
			}

			commitHash, content, err := utils.GenerateHash(reveal)
			if err != nil {
				return err
			}

			// Save reveal file.
			savedBidDir, _ := cmd.Flags().GetString(FlagSaveBidsTo)
			// Make dir if it doesn't exist
			if _, err := os.Stat(savedBidDir); os.IsNotExist(err) {
				if err = os.MkdirAll(savedBidDir, 0o700); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", savedBidDir, err)
				}
			}
			savedBidPath := filepath.Join(savedBidDir, fmt.Sprintf("%s-%s.json", clientCtx.FromName, commitHash))
			err = os.WriteFile(savedBidPath, content, 0o600)
			if err != nil {
				return err
			}

			msg := auctiontypes.NewMsgCommitBid(auctionId, commitHash, clientCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().String(FlagSaveBidsTo, "./bids", "Directory to save bid reveal files")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRevealBid is the CLI command for revealing a bid.
func GetCmdRevealBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal-bid [auction-id] [reveal-file-path]",
		Short: "Reveal bid",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId := args[0]
			revealFilePath := args[1]

			revealBytes, err := os.ReadFile(revealFilePath)
			if err != nil {
				return err
			}

			msg := auctiontypes.NewMsgRevealBid(auctionId, hex.EncodeToString(revealBytes), clientCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetCmdCreateAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [kind] [commits-duration] [reveals-duration] [commit-fee] [reveal-fee] [minimum-bid] [max-price] [num-providers]",
		Short: "Create auction.",
		Args:  cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			kind := args[0]

			commitsDuration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			revealsDuration, err := time.ParseDuration(args[2])
			if err != nil {
				return err
			}

			commitFee, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}

			revealFee, err := sdk.ParseCoinNormalized(args[4])
			if err != nil {
				return err
			}

			minimumBid, err := sdk.ParseCoinNormalized(args[5])
			if err != nil {
				return err
			}

			maxPrice, err := sdk.ParseCoinNormalized(args[6])
			if err != nil {
				return err
			}

			numProviders, err := strconv.ParseInt(args[7], 10, 32)
			if err != nil {
				return err
			}

			msg := auctiontypes.NewMsgCreateAuction(
				kind,
				commitsDuration,
				revealsDuration,
				commitFee,
				revealFee,
				minimumBid,
				maxPrice,
				int32(numProviders),
				clientCtx.GetFromAddress(),
			)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
