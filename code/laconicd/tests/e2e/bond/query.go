package bond

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	"git.vdb.to/cerc-io/laconicd/x/bond/client/cli"
)

func (ets *E2ETestSuite) TestGetQueryBondList() {
	val := ets.network.Validators[0]
	sr := ets.Require()

	testCases := []struct {
		name       string
		args       []string
		createBond bool
		preRun     func()
	}{
		{
			"create and get bond lists",
			[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			func() {
				ets.createBond()
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(fmt.Sprintf("Case %s", tc.name), func() {
			clientCtx := val.ClientCtx
			if tc.createBond {
				tc.preRun()
			}

			cmd := cli.GetQueryBondList()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			sr.NoError(err)
			var queryResponse bondtypes.QueryBondsResponse
			err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &queryResponse)
			sr.NoError(err)
			sr.NotZero(len(queryResponse.GetBonds()))
		})
	}
}
