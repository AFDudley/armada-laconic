package bond

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/x/bond/client/cli"
)

func (ets *E2ETestSuite) TestTxCreateBond() {
	val := ets.network.Validators[0]
	sr := ets.Require()

	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"without deposit",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
			},
			true,
		},
		{
			"create bond",
			[]string{
				fmt.Sprintf("10%s", ets.cfg.BondDenom),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
			},
			false,
		},
	}

	for _, tc := range testCases {
		ets.Run(fmt.Sprintf("Case %s", tc.name), func() {
			clientCtx := val.ClientCtx
			cmd := cli.NewCreateBondCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.err {
				sr.Error(err)
			} else {
				sr.NoError(err)
				var d sdk.TxResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
				sr.Nil(err)
				sr.NoError(err)
				sr.Zero(d.Code)
			}
		})
	}
}
