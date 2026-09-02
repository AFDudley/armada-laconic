package app

import (
	_ "embed"

	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/utils"
)

var (
	//go:embed app.yaml
	AppModuleConfigYAML []byte

	// AppModuleConfig returns the default app config.
	AppModuleConfig = depinject.Configs(
		appconfig.LoadYAML(AppModuleConfigYAML),
		depinject.Provide(
			server.NewGasService,
		),
		depinject.Supply(
			utils.NewAddressCodec,
			utils.NewValidatorAddressCodec,
			utils.NewConsensusAddressCodec,
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName: gov.NewAppModuleBasic(
					[]govclient.ProposalHandler{},
				),
			},
		),
	)
)
