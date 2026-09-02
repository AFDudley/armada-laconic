package cmd

import (
	"errors"
	"io"

	"github.com/cometbft/cometbft/node"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"git.vdb.to/cerc-io/laconicd/app"
	"git.vdb.to/cerc-io/laconicd/gql"
	laconicserver "git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/server/nitro"
	"git.vdb.to/cerc-io/laconicd/server/relay"
)

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	basicManager module.BasicManager,
) *laconicserver.ServerAux {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		NewTestnetCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	addStartComponents := func(startCmd *cobra.Command) {
		laconicserver.SetRequiredComponents(startCmd, &gql.Server{}, &nitro.Server{}, &relay.Server{})
	}
	var srv laconicserver.ServerAux
	server.AddCommandsWithStartCmdOptions(rootCmd, app.DefaultNodeHome, newApp, appExport, server.StartCmdOptions{
		AddFlags: addStartComponents,
		// reactors will be configured at start, after the command is fully initialized
		CometNodeOptions: []node.Option{srv.AddReactors},
	})

	// Capture the genesis command from genutilcli and add new commands
	genesisCmd := genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome)
	genesisCmd.AddCommand(AddGenesisLockupAccountCmd())

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		genesisCommand(txConfig, basicManager),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
	return &srv
}

// initializes all server components needed by a command
func initComponents(
	configMap laconicserver.ConfigMap,
	clientCtx client.Context,
	logger log.Logger,
	needComponent func(string) bool,
) ([]laconicserver.ServerComponent, error) {
	var (
		components  []laconicserver.ServerComponent
		gqlServer   = &gql.Server{}
		nitroServer = &nitro.Server{}
		relayServer = &relay.Server{}
		err         error
	)
	if needComponent(gqlServer.Name()) {
		gqlServer, err = gql.New(clientCtx, configMap, logger.With("module", "gql-server"))
		if err != nil {
			return nil, err
		}
		components = append(components, gqlServer)
	}
	if needComponent(nitroServer.Name()) {
		nitroServer, err = nitro.New(configMap, logger, clientCtx.Keyring)
		if err != nil {
			return nil, err
		}
		components = append(components, nitroServer)
	}
	if needComponent(relayServer.Name()) {
		relayServer, err = relay.New(configMap, logger)
		if err != nil {
			return nil, err
		}
		components = append(components, relayServer)
	}
	return components, nil
}

// genesisCommand builds genesis-related `laconicd genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

// newApp creates the application
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	ret, err := app.NewLaconicApp(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
	if err != nil {
		panic(err)
	}
	return ret
}

// appExport creates a new laconicd (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	var laconicApp *app.LaconicApp
	var err error
	if height != -1 {
		if laconicApp, err = app.NewLaconicApp(logger, db, traceStore, false, appOpts); err != nil {
			return servertypes.ExportedApp{}, err
		}
		if err = laconicApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		if laconicApp, err = app.NewLaconicApp(logger, db, traceStore, true, appOpts); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return laconicApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
