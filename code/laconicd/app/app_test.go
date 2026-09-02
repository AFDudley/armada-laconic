package app

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func TestLaconicAppExportAndBlockedAccounts(t *testing.T) {
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewLaconicAppWithCustomOptions(t, false, SetupOptions{
		Logger:  logger.With("instance", "first"),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// BlockedAddresses returns a map of addresses in app v1 and a map of modules name in app di.
	for _, acc := range BlockedAccounts() {
		addr := app.AccountKeeper.GetModuleAddress(acc)

		require.True(
			t,
			app.BankKeeper.BlockedAddr(addr),
			fmt.Sprintf("ensure that blocked addresses are properly set in bank keeper: %s should be blocked", acc),
		)
	}

	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called
	app2, err := NewLaconicApp(logger.With("instance", "second"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	require.NoError(t, err)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
