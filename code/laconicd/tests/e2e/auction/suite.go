package auction

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	laconictestcli "git.vdb.to/cerc-io/laconicd/testutil/cli"
	"git.vdb.to/cerc-io/laconicd/testutil/network"
	types "git.vdb.to/cerc-io/laconicd/x/auction"
	"git.vdb.to/cerc-io/laconicd/x/auction/client/cli"
)

var (
	ownerAccount  = "owner"
	bidderAccount = "bidder"
	ownerAddress  string
	bidderAddress string
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	defaultAuctionId string
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (ets *E2ETestSuite) SetupSuite() { //nolint: all
	sr := ets.Require()
	ets.T().Log("setting up e2e test suite")

	var err error

	ets.network, err = network.New(ets.T(), ets.T().TempDir(), ets.cfg)
	sr.NoError(err)

	_, err = ets.network.WaitForHeight(1)
	sr.NoError(err)

	// setting up random owner and bidder accounts
	ets.createAccountWithBalance(ownerAccount, &ownerAddress)
	ets.createAccountWithBalance(bidderAccount, &bidderAddress)

	ets.defaultAuctionId = ets.createAuctionAndBid(true, false)
}

func (ets *E2ETestSuite) TearDownSuite() {
	ets.T().Log("tearing down integration test suite")
	ets.network.Cleanup()

	ets.cleanupBidFiles()
}

func (ets *E2ETestSuite) createAccountWithBalance(accountName string, accountAddress *string) {
	val := ets.network.Validators[0]
	sr := ets.Require()

	info, _, err := val.ClientCtx.Keyring.NewMnemonic(accountName, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	sr.NoError(err)

	newAddr, _ := info.GetAddress()
	out, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(200000))),
		addresscodec.NewBech32Codec("laconic"),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(10))).String()),
	)
	sr.NoError(err)

	var response sdk.TxResponse
	sr.NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, response.TxHash, 0))

	*accountAddress = newAddr.String()
}

func (ets *E2ETestSuite) createAuctionAndBid(createAuction, createBid bool) string {
	val := ets.network.Validators[0]
	sr := ets.Require()
	auctionId := ""

	if createAuction {
		auctionArgs := []string{
			types.AuctionKindVickrey,
			sampleCommitTime, sampleRevealTime,
			fmt.Sprintf("10%s", ets.cfg.BondDenom),
			fmt.Sprintf("10%s", ets.cfg.BondDenom),
			fmt.Sprintf("100%s", ets.cfg.BondDenom),
			fmt.Sprintf("0%s", ets.cfg.BondDenom),
			"0",
		}

		resp, err := ets.executeTx(cli.GetCmdCreateAuction(), auctionArgs, ownerAccount)
		sr.NoError(err)
		sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, resp.TxHash, 0))

		out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.GetCmdList(), queryJSONFlag)
		sr.NoError(err)
		var queryResponse types.QueryAuctionsResponse
		err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &queryResponse)
		sr.NoError(err)
		auctionId = queryResponse.Auctions.Auctions[0].Id
	} else {
		auctionId = ets.defaultAuctionId
	}

	if createBid {
		bidArgs := []string{auctionId, fmt.Sprintf("200%s", ets.cfg.BondDenom)}
		resp, err := ets.executeTx(cli.GetCmdCommitBid(), bidArgs, bidderAccount)
		sr.NoError(err)
		sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, resp.TxHash, 0))
	}

	return auctionId
}

func (ets *E2ETestSuite) cleanupBidFiles() {
	matches, err := filepath.Glob(fmt.Sprintf("%s-*.json", bidderAccount))
	if err != nil {
		ets.T().Errorf("Error matching bidder files: %v\n", err)
	}

	for _, match := range matches {
		err := os.Remove(match)
		if err != nil {
			ets.T().Errorf("Error removing bidder file: %v\n", err)
		}
	}
}
