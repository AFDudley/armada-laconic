package e2e

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/codec"

	dbm "github.com/cosmos/cosmos-db"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	laconicApp "git.vdb.to/cerc-io/laconicd/app"
	auctionmodule "git.vdb.to/cerc-io/laconicd/x/auction/module"
	bondmodule "git.vdb.to/cerc-io/laconicd/x/bond/module"
	registrymodule "git.vdb.to/cerc-io/laconicd/x/registry/module"

	"git.vdb.to/cerc-io/laconicd/app/params"
	_ "git.vdb.to/cerc-io/laconicd/app/params" // import for side-effects (see init)
	"git.vdb.to/cerc-io/laconicd/testutil/network"
)

// NewTestNetworkFixture returns a new LaconicApp AppConstructor for network simulation tests
func NewTestNetworkFixture() network.TestFixture {
	dir, err := os.MkdirTemp("", "laconic")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)

	app, err := laconicApp.NewLaconicApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(dir))
	if err != nil {
		panic(fmt.Sprintf("failed to create laconic app: %v", err))
	}

	appCtr := func(val network.ValidatorI) servertypes.Application {
		app, err := laconicApp.NewLaconicApp(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			bam.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
		)
		if err != nil {
			panic(fmt.Sprintf("failed creating temporary directory: %v", err))
		}

		return app
	}

	encodingConfig := testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		auctionmodule.AppModule{},
		bondmodule.AppModule{},
		registrymodule.AppModule{},
	)

	genesisState := app.DefaultGenesis()
	genesisState, err = updateStakingGenesisBondDenom(genesisState, encodingConfig.Codec)
	if err != nil {
		panic(fmt.Sprintf("failed to update genesis state: %v", err))
	}

	return network.TestFixture{
		AppConstructor: appCtr,
		GenesisState:   genesisState,
		EncodingConfig: encodingConfig,
	}
}

func updateStakingGenesisBondDenom(genesisState map[string]json.RawMessage, codec codec.Codec) (map[string]json.RawMessage, error) {
	var stakingGenesis stakingtypes.GenesisState
	if err := codec.UnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesis); err != nil {
		return nil, nil
	}
	stakingGenesis.Params.BondDenom = params.CoinUnit

	stakingGenesisBz, err := codec.MarshalJSON(&stakingGenesis)
	if err != nil {
		return nil, nil
	}
	genesisState[stakingtypes.ModuleName] = stakingGenesisBz

	return genesisState, nil
}
