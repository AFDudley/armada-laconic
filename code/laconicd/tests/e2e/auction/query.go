package auction

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	types "git.vdb.to/cerc-io/laconicd/x/auction"
	"git.vdb.to/cerc-io/laconicd/x/auction/client/cli"
)

var queryJSONFlag = []string{fmt.Sprintf("--%s=json", flags.FlagOutput)}

func (ets *E2ETestSuite) TestGetCmdList() {
	val := ets.network.Validators[0]
	sr := ets.Require()

	testCases := []struct {
		msg           string
		createAuction bool
	}{
		{
			"list auctions when no auctions exist",
			false,
		},
		{
			"list auctions after creating an auction",
			true,
		},
	}

	for _, test := range testCases {
		ets.Run(fmt.Sprintf("Case %s", test.msg), func() {
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.GetCmdList(), queryJSONFlag)
			sr.NoError(err)
			var auctions types.QueryAuctionsResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &auctions)
			sr.NoError(err)
			if test.createAuction {
				sr.NotZero(len(auctions.Auctions.Auctions))
			}
		})
	}
}
