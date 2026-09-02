package keeper_test

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/app/params"
	integrationTest "git.vdb.to/cerc-io/laconicd/tests/integration"
	types "git.vdb.to/cerc-io/laconicd/x/auction"
)

const testCommitHash = "71D8CF34026E32A3A34C2C2D4ADF25ABC8D7943A4619761BE27F196603D91B9D"

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

func (kts *KeeperTestSuite) TestGrpcGetAuction() {
	testCases := []struct {
		msg           string
		req           *types.QueryGetAuctionRequest
		createAuction bool
	}{
		{
			"fetch auction with empty auction ID",
			&types.QueryGetAuctionRequest{},
			false,
		},
		{
			"fetch auction with valid auction ID",
			&types.QueryGetAuctionRequest{},
			true,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			var expectedAuction types.Auction
			if test.createAuction {
				auction, _, err := kts.createAuctionAndCommitBid(false)
				kts.Require().NoError(err)
				test.req.Id = auction.Id
				expectedAuction = *auction
			}

			resp, err := kts.queryClient.GetAuction(context.Background(), test.req)
			if test.createAuction {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.GetAuction())
				kts.Require().EqualExportedValues(expectedAuction, *(resp.GetAuction()))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetAllAuctions() {
	testCases := []struct {
		msg            string
		req            *types.QueryAuctionsRequest
		createAuctions bool
		auctionCount   int
	}{
		{
			"fetch auctions when no auctions exist",
			&types.QueryAuctionsRequest{},
			false,
			0,
		},

		{
			"fetch auctions with one auction created",
			&types.QueryAuctionsRequest{},
			true,
			1,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuctions {
				_, _, err := kts.createAuctionAndCommitBid(false)
				kts.Require().NoError(err)
			}

			resp, _ := kts.queryClient.Auctions(context.Background(), test.req)
			kts.Require().Equal(test.auctionCount, len(resp.GetAuctions().Auctions))
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetBids() {
	testCases := []struct {
		msg           string
		req           *types.QueryGetBidsRequest
		createAuction bool
		commitBid     bool
		bidCount      int
	}{
		{
			"fetch all bids when no auction exists",
			&types.QueryGetBidsRequest{},
			false,
			false,
			0,
		},
		{
			"fetch all bids for valid auction but no added bids",
			&types.QueryGetBidsRequest{},
			true,
			false,
			0,
		},
		{
			"fetch all bids for valid auction and valid bid",
			&types.QueryGetBidsRequest{},
			true,
			true,
			1,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuction {
				auction, _, err := kts.createAuctionAndCommitBid(test.commitBid)
				kts.Require().NoError(err)
				test.req.AuctionId = auction.Id
			}

			resp, err := kts.queryClient.GetBids(context.Background(), test.req)
			if test.createAuction {
				kts.Require().NoError(err)
				kts.Require().Equal(test.bidCount, len(resp.GetBids()))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetBid() {
	testCases := []struct {
		msg                 string
		req                 *types.QueryGetBidRequest
		createAuctionAndBid bool
	}{
		{
			"fetch bid when bid does not exist",
			&types.QueryGetBidRequest{},
			false,
		},
		{
			"fetch bid when valid bid exists",
			&types.QueryGetBidRequest{},
			true,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuctionAndBid {
				auction, bid, err := kts.createAuctionAndCommitBid(test.createAuctionAndBid)
				kts.Require().NoError(err)
				test.req.AuctionId = auction.Id
				test.req.Bidder = bid.BidderAddress
			}

			resp, err := kts.queryClient.GetBid(context.Background(), test.req)
			if test.createAuctionAndBid {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.Bid)
				kts.Require().Equal(test.req.Bidder, resp.Bid.BidderAddress)
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetAuctionsByBidder() {
	testCases := []struct {
		msg                       string
		req                       *types.QueryAuctionsByBidderRequest
		createAuctionAndCommitBid bool
		auctionCount              int
	}{
		{
			"get auctions by bidder with invalid bidder address",
			&types.QueryAuctionsByBidderRequest{},
			false,
			0,
		},
		{
			"get auctions by bidder with valid auction and bid",
			&types.QueryAuctionsByBidderRequest{},
			true,
			1,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuctionAndCommitBid {
				_, bid, err := kts.createAuctionAndCommitBid(test.createAuctionAndCommitBid)
				kts.Require().NoError(err)
				test.req.BidderAddress = bid.BidderAddress
			}

			resp, err := kts.queryClient.AuctionsByBidder(context.Background(), test.req)
			if test.createAuctionAndCommitBid {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.Auctions)
				kts.Require().Equal(test.auctionCount, len(resp.Auctions.Auctions))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetAuctionsByOwner() {
	testCases := []struct {
		msg           string
		req           *types.QueryAuctionsByOwnerRequest
		createAuction bool
		auctionCount  int
	}{
		{
			"get auctions by owner with invalid owner address",
			&types.QueryAuctionsByOwnerRequest{},
			false,
			0,
		},
		{
			"get auctions by owner with valid auction",
			&types.QueryAuctionsByOwnerRequest{},
			true,
			1,
		},
	}

	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuction {
				auction, _, err := kts.createAuctionAndCommitBid(false)
				kts.Require().NoError(err)
				test.req.OwnerAddress = auction.OwnerAddress
			}

			resp, err := kts.queryClient.AuctionsByOwner(context.Background(), test.req)
			if test.createAuction {
				kts.Require().NoError(err)
				kts.Require().NotNil(resp.Auctions)
				kts.Require().Equal(test.auctionCount, len(resp.Auctions.Auctions))
			} else {
				kts.Require().NotNil(err)
				kts.Require().Error(err)
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcQueryBalance() {

	testCases := []struct {
		msg           string
		req           *types.QueryGetAuctionModuleBalanceRequest
		createAuction bool
		auctionCount  int
	}{
		{
			"get balance with no auctions created",
			&types.QueryGetAuctionModuleBalanceRequest{},
			false,
			0,
		},
		{
			"get balance with single auction created",
			&types.QueryGetAuctionModuleBalanceRequest{},
			true,
			1,
		},
	}

	for _, test := range testCases {
		if test.createAuction {
			_, _, err := kts.createAuctionAndCommitBid(true)
			kts.Require().NoError(err)
		}

		resp, err := kts.queryClient.GetAuctionModuleBalance(context.Background(), test.req)
		kts.Require().NoError(err)
		kts.Require().Equal(test.auctionCount, len(resp.GetBalance()))
	}
}

func (kts *KeeperTestSuite) createAuctionAndCommitBid(commitBid bool) (*types.Auction, *types.Bid, error) {
	ctx, k := kts.SdkCtx, kts.AuctionKeeper
	accCount := 1
	if commitBid {
		accCount++
	}

	// Create funded account(s)
	accounts := simtestutil.AddTestAddrs(kts.BankKeeper, integrationTest.BondDenomProvider{}, ctx, accCount, math.NewInt(1000000))

	auction, err := k.CreateAuction(
		ctx,
		types.MsgCreateAuction{
			Kind:            types.AuctionKindVickrey,
			Signer:          accounts[0].String(),
			CommitsDuration: 5 * time.Minute,
			RevealsDuration: 5 * time.Minute,
			CommitFee:       sdk.NewCoin(params.CoinUnit, sdkmath.NewInt(1000)),
			RevealFee:       sdk.NewCoin(params.CoinUnit, sdkmath.NewInt(1000)),
			MinimumBid:      sdk.NewCoin(params.CoinUnit, sdkmath.NewInt(1000000)),
		},
	)
	if err != nil {
		return nil, nil, err
	}

	if commitBid {
		bid, err := k.CommitBid(ctx, types.NewMsgCommitBid(auction.Id, testCommitHash, accounts[1]))
		if err != nil {
			return nil, nil, err
		}

		return auction, bid, nil
	}

	return auction, nil, nil
}
