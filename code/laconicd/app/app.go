package app

import (
	_ "embed"
	"io"

	dbm "github.com/cosmos/cosmos-db"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	auctionkeeper "git.vdb.to/cerc-io/laconicd/x/auction/keeper"
	bondkeeper "git.vdb.to/cerc-io/laconicd/x/bond/keeper"
	nitrokeeper "git.vdb.to/cerc-io/laconicd/x/nitro/keeper"
	onboardingkeeper "git.vdb.to/cerc-io/laconicd/x/onboarding/keeper"
	registrykeeper "git.vdb.to/cerc-io/laconicd/x/registry/keeper"
	types "git.vdb.to/cerc-io/laconicd/x/types/v1"

	_ "cosmossdk.io/api/cosmos/tx/config/v1"          // import for side-effects
	_ "cosmossdk.io/x/evidence"                       // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"      // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/mint"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/protocolpool"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/slashing"       // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"        // import for side-effects

	_ "git.vdb.to/cerc-io/laconicd/x/auction/module"    // import for side-effects
	_ "git.vdb.to/cerc-io/laconicd/x/bond/module"       // import for side-effects
	_ "git.vdb.to/cerc-io/laconicd/x/onboarding/module" // import for side-effects
	_ "git.vdb.to/cerc-io/laconicd/x/registry/module"   // import for side-effects
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var (
	_ runtime.AppI            = (*LaconicApp)(nil)
	_ servertypes.Application = (*LaconicApp)(nil)
)

// LaconicApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type LaconicApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// essential keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper

	// supplementary keepers
	ProtocolPoolKeeper protocolpoolkeeper.Keeper

	// laconic keepers
	AuctionKeeper    *auctionkeeper.Keeper
	BondKeeper       *bondkeeper.Keeper
	RegistryKeeper   registrykeeper.Keeper
	OnboardingKeeper *onboardingkeeper.Keeper
	NitroKeeper      nitrokeeper.Keeper

	// simulation manager
	sm *module.SimulationManager
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".laconicd")
	if err != nil {
		panic(err)
	}
}

// NewLaconicApp returns a reference to an initialized LaconicApp.
func NewLaconicApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*LaconicApp, error) {
	var (
		app        = &LaconicApp{}
		appBuilder *runtime.AppBuilder

		appConfig = depinject.Configs(
			AppModuleConfig,
			depinject.Supply(
				logger,
				appOpts,
			),
		)
	)

	// if err := depinject.Inject(appConfig,
	if err := depinject.InjectDebug(
		depinject.FileLogger("/Users/roy/vulcanize/dump/laconic-debug/depinject_app.log"),
		appConfig,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.ConsensusParamsKeeper,
		&app.EvidenceKeeper,
		&app.ProtocolPoolKeeper,
		&app.AuctionKeeper,
		&app.BondKeeper,
		&app.RegistryKeeper,
		&app.OnboardingKeeper,
		&app.NitroKeeper,
	); err != nil {
		return nil, err
	}

	// Register custom interfaces
	RegisterCustomInterfaces(app.interfaceRegistry)
	RegisterCustomLegacyAminoCodec(app.legacyAmino)

	// create and set vote extension handler
	voteExtOp := func(bApp *baseapp.BaseApp) {
		voteExtHandler := app.NewVoteExtensionHandler()
		voteExtHandler.SetHandlers(bApp)
	}
	baseAppOptions = append(baseAppOptions, voteExtOp, baseapp.SetOptimisticExecution())

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	// register streaming services
	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	/****  Module Options ****/

	// create the simulation manager and define the order of the modules for deterministic simulations
	// NOTE: this is not required apps that don't use the simulator for fuzz testing transactions
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, make(map[string]module.AppModuleSimulation, 0))
	app.sm.RegisterStoreDecoders()

	// set custom ante handler
	app.setAnteHandler()

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

// setCustomAnteHandler overwrites default ante handlers with custom ante handlers
// Reference: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/tx/config/config.go#L149
func (app *LaconicApp) setAnteHandler() error {
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: app.txConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			TxFeeChecker:    checkTxFeeWithValidatorMinGasPrices,
		},
	)
	if err != nil {
		return err
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
	return nil
}

// LegacyAmino returns LaconicApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *LaconicApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns LaconicApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *LaconicApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns LaconicApp's InterfaceRegistry.
func (app *LaconicApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns LaconicApp's TxConfig
func (app *LaconicApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *LaconicApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	sk := app.UnsafeFindStoreKey(storeKey)
	kvStoreKey, ok := sk.(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

func (app *LaconicApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

// SimulationManager implements the SimulationApp interface
func (app *LaconicApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *LaconicApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func RegisterCustomInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	// Custom LockupAccount type needs to be registered to be able to use it as a genesis account
	interfaceRegistry.RegisterImplementations((*authtypes.GenesisAccount)(nil), &types.LockupAccount{})

	// LockupAccount extends auth Account
	interfaceRegistry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*sdk.AccountI)(nil),
		&types.LockupAccount{},
	)

	// LockupAccount extends auth ModuleAccount
	interfaceRegistry.RegisterInterface(
		"cosmos.auth.v1beta1.ModuleAccountI",
		(*sdk.ModuleAccountI)(nil),
		&types.LockupAccount{},
	)

	interfaceRegistry.RegisterInterface(
		"cosmos.auth.v1beta1.GenesisAccount",
		(*authtypes.GenesisAccount)(nil),
		&types.LockupAccount{},
	)

	interfaceRegistry.RegisterInterface(
		"cerc.types.v1.LockupAccountI",
		(*types.LockupAccountI)(nil),
		&types.LockupAccount{},
	)
}

func RegisterCustomLegacyAminoCodec(registrar *codec.LegacyAmino) {
	registrar.RegisterInterface((*types.LockupAccountI)(nil), nil)
	registrar.RegisterConcrete(&types.LockupAccount{}, "laconic/LockupAccount", nil)
}

// setCustomAnteHandler overwrites default ante handlers with custom ante handlers
// Reference: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/tx/config/config.go#L149
func (app *LaconicApp) setCustomAnteHandler() {
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: app.txConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			TxFeeChecker:    checkTxFeeWithValidatorMinGasPrices,
		},
	)
	if err != nil {
		panic(err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}
