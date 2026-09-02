package registry

import (
	"fmt"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/x/registry/client/cli"
)

func (ets *E2ETestSuite) TestGetCmdSetRecord() {
	val := ets.network.Validators[0]
	sr := ets.Require()

	bondId := ets.bondId
	payloadPath := "../../data/examples/service_provider_example.yml"
	payloadFilePath, err := filepath.Abs(payloadPath)
	sr.NoError(err)

	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"invalid request without bond id/without payload",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
			},
			true,
		},
		{
			"success",
			[]string{
				payloadFilePath, bondId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
			},
			false,
		},
	}

	for _, tc := range testCases {
		ets.Run(fmt.Sprintf("Case %s", tc.name), func() {
			clientCtx := val.ClientCtx
			cmd := cli.GetCmdSetRecord()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.err {
				sr.Error(err)
			} else {
				sr.NoError(err)
				var d sdk.TxResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
				sr.NoError(err)
				sr.Zero(d.Code)
			}
		})
	}
}
