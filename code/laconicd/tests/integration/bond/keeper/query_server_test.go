package keeper_test

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/app/params"
	integrationTest "git.vdb.to/cerc-io/laconicd/tests/integration"
	types "git.vdb.to/cerc-io/laconicd/x/bond"
)

func (kts *KeeperTestSuite) TestGrpcQueryParams() {
	testCases := []struct {
		msg string
		req *types.QueryParamsRequest
	}{
		{
			"fetch params",
			&types.QueryParamsRequest{},
		},
	}
	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			resp, err := kts.queryClient.Params(context.Background(), test.req)
			kts.Require().NoError(err)
			kts.Require().Equal(*(resp.Params), types.DefaultParams())
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcQueryBondsList() {
	testCases := []struct {
		msg         string
		req         *types.QueryBondsRequest
		resp        *types.QueryBondsResponse
		noOfBonds   int
		createBonds bool
	}{
		{
			"empty request",
			&types.QueryBondsRequest{},
			&types.QueryBondsResponse{},
			0,
			false,
		},
		{
			"Get Bonds",
			&types.QueryBondsRequest{},
			&types.QueryBondsResponse{},
			1,
			true,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createBonds {
				_, err := kts.createBond()
				kts.Require().NoError(err)
			}
			resp, _ := kts.queryClient.Bonds(context.Background(), test.req)
			kts.Require().Equal(test.noOfBonds, len(resp.GetBonds()))
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcQueryBondByBondId() {
	testCases := []struct {
		msg         string
		req         *types.QueryGetBondByIdRequest
		createBonds bool
		errResponse bool
		bondId      string
	}{
		{
			"empty request",
			&types.QueryGetBondByIdRequest{},
			false,
			true,
			"",
		},
		{
			"Get Bond By ID",
			&types.QueryGetBondByIdRequest{},
			true,
			false,
			"",
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createBonds {
				bond, err := kts.createBond()
				kts.Require().NoError(err)
				test.req.Id = bond.Id
			}
			resp, err := kts.queryClient.GetBondById(context.Background(), test.req)
			if !test.errResponse {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.GetBond())
				kts.Require().Equal(test.req.Id, resp.GetBond().GetId())
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetBondsByOwner() {
	testCases := []struct {
		msg         string
		req         *types.QueryGetBondsByOwnerRequest
		noOfBonds   int
		createBonds bool
		errResponse bool
		bondId      string
	}{
		{
			"empty request",
			&types.QueryGetBondsByOwnerRequest{},
			0,
			false,
			true,
			"",
		},
		{
			"Get Bond By Owner",
			&types.QueryGetBondsByOwnerRequest{},
			1,
			true,
			false,
			"",
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createBonds {
				bond, err := kts.createBond()
				kts.Require().NoError(err)
				test.req.Owner = bond.Owner
			}
			resp, err := kts.queryClient.GetBondsByOwner(context.Background(), test.req)
			if !test.errResponse {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.GetBonds())
				kts.Require().Equal(test.noOfBonds, len(resp.GetBonds()))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetModuleBalance() {
	testCases := []struct {
		msg         string
		req         *types.QueryGetBondModuleBalanceRequest
		noOfBonds   int
		createBonds bool
		errResponse bool
	}{
		{
			"empty request",
			&types.QueryGetBondModuleBalanceRequest{},
			0,
			true,
			false,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createBonds {
				_, err := kts.createBond()
				kts.Require().NoError(err)
			}
			resp, err := kts.queryClient.GetBondModuleBalance(context.Background(), test.req)
			if !test.errResponse {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.GetBalance())
				kts.Require().Equal(resp.GetBalance(), sdk.NewCoins(sdk.NewCoin(params.CoinUnit, math.NewInt(10))))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) createBond() (*types.Bond, error) {
	ctx, k := kts.SdkCtx, kts.BondKeeper
	accCount := 1

	// Create funded account(s)
	accounts := simtestutil.AddTestAddrs(kts.BankKeeper, integrationTest.BondDenomProvider{}, ctx, accCount, math.NewInt(1000))

	bond, err := k.CreateBond(ctx, accounts[0], sdk.NewCoins(sdk.NewCoin(params.CoinUnit, math.NewInt(10))))
	if err != nil {
		return nil, err
	}

	return bond, nil
}
