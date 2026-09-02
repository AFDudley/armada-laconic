package auction

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	laconictestcli "git.vdb.to/cerc-io/laconicd/testutil/cli"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
	"git.vdb.to/cerc-io/laconicd/x/auction/client/cli"
)

const (
	sampleCommitTime     = "90s"
	sampleRevealTime     = "5s"
	placeholderAuctionId = "placeholder_auction_id"
)

func (ets *E2ETestSuite) TestTxCommitBid() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	testCases := []struct {
		msg           string
		args          []string
		createAuction bool
	}{
		{
			"commit bid with missing args",
			[]string{fmt.Sprintf("200%s", ets.cfg.BondDenom)},
			false,
		},
		{
			"commit bid with valid args",
			[]string{
				placeholderAuctionId,
				fmt.Sprintf("200%s", ets.cfg.BondDenom),
			},
			true,
		},
	}

	for _, test := range testCases {
		ets.Run(fmt.Sprintf("Case %s", test.msg), func() {
			if test.createAuction {
				auctionArgs := []string{
					auctiontypes.AuctionKindVickrey,
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

				out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.GetCmdList(),
					[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)})
				sr.NoError(err)
				var queryResponse auctiontypes.QueryAuctionsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &queryResponse)
				sr.NoError(err)
				sr.NotNil(queryResponse.GetAuctions())
				test.args[0] = queryResponse.GetAuctions().Auctions[0].Id
			}

			resp, err := ets.executeTx(cli.GetCmdCommitBid(), test.args, bidderAccount)
			if test.createAuction {
				sr.NoError(err)
				sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, resp.TxHash, 0))
			} else {
				sr.Error(err)
			}
		})
	}
}

func (ets *E2ETestSuite) executeTx(cmd *cobra.Command, args []string, caller string) (sdk.TxResponse, error) {
	val := ets.network.Validators[0]
	additionalArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, caller),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	args = append(args, additionalArgs...)

	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	var resp sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return resp, nil
}
