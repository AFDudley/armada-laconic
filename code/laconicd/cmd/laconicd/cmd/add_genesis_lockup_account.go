package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	types "git.vdb.to/cerc-io/laconicd/x/types/v1"
)

const (
	flagAppendMode = "append"
)

// AddGenesisLockupAccountCmd returns add-genesis-lockup-account cobra Command.
func AddGenesisLockupAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-lockup-account <account_name> <distribution-json-file> <coin>[,<coin>...]",
		Short: "Add genesis lockup account with give name",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			moduleName := args[0]
			distributionFilePath := args[1]

			var distribution []interface{}
			distributionBytes, err := os.ReadFile(distributionFilePath)
			if err != nil {
				return fmt.Errorf("failed to read %s file: %w", distributionFilePath, err)
			}
			if err = json.Unmarshal(distributionBytes, &distribution); err != nil {
				return fmt.Errorf("distribution is invalid json: %v", err)
			}

			appendflag, _ := cmd.Flags().GetBool(flagAppendMode)

			return AddGenesisLockupAccount(clientCtx.Codec, moduleName, string(distributionBytes), appendflag, config.GenesisFile(), args[2])
		},
	}

	cmd.Flags().Bool(flagAppendMode, false, "append the coins to an account already in the genesis.json file")

	return cmd
}

func AddGenesisLockupAccount(
	cdc codec.Codec,
	moduleName string,
	distribution string,
	appendAcct bool,
	genesisFileURL, amountStr string,
) error {
	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	// create concrete account type based on input parameters
	moduleAccount := authtypes.NewEmptyModuleAccount(moduleName)
	accAddr := moduleAccount.GetAddress()
	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}

	genAccount := &types.LockupAccount{
		BaseAccount:  moduleAccount.BaseAccount,
		Name:         moduleAccount.Name,
		Distribution: distribution,
	}

	if err := genAccount.Validate(); err != nil {
		return fmt.Errorf("failed to validate new genesis account: %w", err)
	}

	appState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genesisFileURL)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
	if accs.Contains(accAddr) {
		if !appendAcct {
			return fmt.Errorf(" Account %s already exists\nUse `append` flag to append account at existing address", accAddr)
		}

		genesisB := banktypes.GetGenesisStateFromAppState(cdc, appState)
		for idx, acc := range genesisB.Balances {
			if acc.Address != accAddr.String() {
				continue
			}

			updatedCoins := acc.Coins.Add(coins...)
			bankGenState.Balances[idx] = banktypes.Balance{Address: accAddr.String(), Coins: updatedCoins.Sort()}
			break
		}
	} else {
		// Add the new account to the set of genesis accounts and sanitize the accounts afterwards.
		accs = append(accs, genAccount)
		accs = authtypes.SanitizeGenesisAccounts(accs)

		genAccs, err := authtypes.PackAccounts(accs)
		if err != nil {
			return fmt.Errorf("failed to convert accounts into any's: %w", err)
		}
		authGenState.Accounts = genAccs

		authGenStateBz, err := cdc.MarshalJSON(&authGenState)
		if err != nil {
			return fmt.Errorf("failed to marshal auth genesis state: %w", err)
		}
		appState[authtypes.ModuleName] = authGenStateBz

		bankGenState.Balances = append(bankGenState.Balances, balances)
	}

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenState.Supply = bankGenState.Supply.Add(balances.Coins...)

	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	appGenesis.AppState = appStateJSON
	return genutil.ExportGenesisFile(appGenesis, genesisFileURL)
}
