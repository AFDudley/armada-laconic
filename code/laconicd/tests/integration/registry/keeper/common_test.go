package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"git.vdb.to/cerc-io/laconicd/app/params"
	integrationTest "git.vdb.to/cerc-io/laconicd/tests/integration"
	bondTypes "git.vdb.to/cerc-io/laconicd/x/bond"
	types "git.vdb.to/cerc-io/laconicd/x/registry"
)

type KeeperTestSuite struct {
	suite.Suite
	integrationTest.TestFixture

	queryClient types.QueryClient

	accounts []sdk.AccAddress
	bond     bondTypes.Bond
}

func (kts *KeeperTestSuite) SetupTest() {
	err := kts.TestFixture.Setup()
	kts.Require().NoError(err)

	// set default params
	err = kts.RegistryKeeper.Params.Set(kts.SdkCtx, types.DefaultParams())
	kts.Require().NoError(err)

	qr := kts.App.QueryHelper()
	kts.queryClient = types.NewQueryClient(qr)

	// Create a bond
	bond, err := kts.createBond()
	kts.Require().NoError(err)
	kts.bond = *bond
}

func TestRegistryKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (kts *KeeperTestSuite) createBond() (*bondTypes.Bond, error) {
	ctx := kts.SdkCtx

	// Create a funded account
	kts.accounts = simtestutil.AddTestAddrs(kts.BankKeeper, integrationTest.BondDenomProvider{}, ctx, 1, math.NewInt(1000000000000))

	bond, err := kts.BondKeeper.CreateBond(ctx, kts.accounts[0], sdk.NewCoins(sdk.NewCoin(params.CoinUnit, math.NewInt(1000000000))))
	if err != nil {
		return nil, err
	}

	return bond, nil
}
