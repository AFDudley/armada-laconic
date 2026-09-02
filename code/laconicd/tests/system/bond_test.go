package system

import (
	"encoding/json"
	"testing"

	"git.vdb.to/cerc-io/laconicd/x/bond"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type bondSuite struct {
	testSuite
}

func TestBonds(t *testing.T) {
	suite.Run(t, new(bondSuite))
}

func (s *bondSuite) SetupTest() {
	s.resetChain()
	Sut.StartChain(s.T())
}

func (s *bondSuite) TestQueryList() {
	testCases := []struct {
		name   string
		args   []string
		prerun func()
	}{
		{
			"create and get bond lists",
			nil,
			func() { s.createBond() },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prerun()
			out := s.cli().CustomQuery("q", "bond", "list")
			s.Require().NotEmpty(gjson.Get(out, "bonds").Array(), out)
		})
	}
}

func (s *bondSuite) TestTxCreateBond() {
	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"without deposit",
			[]string{},
			true,
		},
		{
			"create bond",
			[]string{
				"10" + s.bondDenom,
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := append([]string{"tx", "bond", "create"}, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *bondSuite) createBond() string {
	return createBond(&s.testSuite)
}

// free function for use in other suites
func createBond(s *testSuite) string {
	require := s.Require()

	cli := s.cli()
	cmd := []string{
		"tx", "bond", "create", ("1000000" + s.bondDenom),
	}
	s.runTx(cmd, false)

	raw := cli.CustomQuery("q", "bond", "list")
	var queryResponse bond.QueryBondsResponse
	require.NoError(json.Unmarshal([]byte(raw), &queryResponse))
	bonds := queryResponse.GetBonds()
	require.NotEmpty(bonds)

	return bonds[0].GetId()
}
