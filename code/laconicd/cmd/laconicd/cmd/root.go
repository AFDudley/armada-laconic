package cmd

import (
	"os"

	"github.com/spf13/cobra"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"git.vdb.to/cerc-io/laconicd/app"
	appconfig "git.vdb.to/cerc-io/laconicd/app/config"
)

const (
	// EnvPrefix is the environment variable prefix for the application
	EnvPrefix = "LACONIC"

	// DefaultNodeHome is the default data directory name for the application
	DefaultNodeHome = ".laconicd"
)

// NewRootCmd creates a new root command for laconicd. It is called once in the main function.
func NewRootCmd(args ...string) (*cobra.Command, error) {
	var (
		txConfigOpts       tx.ConfigOptions
		autoCliOpts        autocli.AppOptions
		moduleBasicManager module.BasicManager
		clientCtx          client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(app.AppModuleConfig,
			depinject.Supply(
				log.NewNopLogger(),
			),
			depinject.Provide(
				ProvideClientContext,
			),
		),
		&txConfigOpts,
		&autoCliOpts,
		&moduleBasicManager,
		&clientCtx,
	); err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use:               "laconicd",
		SilenceErrors:     true,
		SilenceUsage:      true, // prevent usage printing on every error
		PersistentPreRunE: RootCommandPersistentPreRun(clientCtx, txConfigOpts),
	}

	initCtx := initRootCmd(rootCmd, clientCtx.TxConfig, moduleBasicManager)

	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		return nil, err
	}
	rootCmd.SetArgs(args)

	// now enhance the subcommand
	subCmd, _, err := rootCmd.Find(args)
	if err != nil {
		return nil, err
	}
	initCtx.AddComponents(subCmd, initComponents)
	return rootCmd, nil
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfig client.TxConfig,
	legacyAmino *codec.LegacyAmino,
) client.Context {
	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(EnvPrefix) // env variable prefix

	// Read the config again to overwrite the default values with the values from the config file
	clientCtx, _ = clientconfig.ReadFromClientConfig(clientCtx)

	// Workaround: Unset clientCtx.HomeDir and clientCtx.KeyringDir from depinject clientCtx as they are given precedence over
	// the CLI args (--home flag) in some commands
	// TODO: Implement proper fix
	clientCtx.HomeDir = ""
	clientCtx.KeyringDir = ""

	// XXX TODO fix after rebase
	// Custom LockupAccount type needs to be registered
	// interfaceRegistry.RegisterImplementations((*types.LockupAccountI)(nil), &types.LockupAccount{})
	// interfaceRegistry.RegisterImplementations((*authtypes.GenesisAccount)(nil), &types.LockupAccount{})

	return clientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return keyring.NewAutoCLIKeyring(kb)
}

func RootCommandPersistentPreRun(
	clientCtx client.Context,
	txConfigOpts tx.ConfigOptions,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		// set the default command outputs
		cmd.SetOut(cmd.OutOrStdout())
		cmd.SetErr(cmd.ErrOrStderr())

		clientCtx = clientCtx.WithCmdContext(cmd.Context()).WithViper("")
		clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		clientCtx, err = clientconfig.ReadFromClientConfig(clientCtx)
		if err != nil {
			return err
		}

		// This needs to go after ReadFromClientConfig, as that function
		// sets the RPC client needed for SIGN_MODE_TEXTUAL. This sign mode
		// is only available if the client is online.
		if !clientCtx.Offline {
			txConfigOpts.EnabledSignModes = append(txConfigOpts.EnabledSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts.TextualCoinMetadataQueryFn = txmodule.NewGRPCCoinMetadataQueryFn(clientCtx)
			txConfigWithTextual, err := tx.NewTxConfigWithOptions(codec.NewProtoCodec(clientCtx.InterfaceRegistry), txConfigOpts)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithTxConfig(txConfigWithTextual)
		}

		if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
			return err
		}

		appConfig := appconfig.DefaultConfig()
		cmtConfig := initCometBFTConfig()

		serverCtx, err := server.InterceptConfigsAndCreateContext(
			cmd, appconfig.DefaultConfigTemplate, appConfig, cmtConfig,
		)
		if err != nil {
			return err
		}
		// use slog-based logger to combine with nitro logs
		logger, err := createSlogLogger(serverCtx.Viper.AllSettings(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		serverCtx.Logger = logger.With(log.ModuleKey, "server")

		return server.SetCmdServerContext(cmd, serverCtx)
	}
}
