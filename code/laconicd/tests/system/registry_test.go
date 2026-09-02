package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	// accounts "github.com/cosmos/cosmos-sdk/x/accounts/v1"
	"cosmossdk.io/math"
	systest "cosmossdk.io/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// nitro "git.vdb.to/cerc-io/laconicd/x/nitro/v1"
	"git.vdb.to/cerc-io/laconicd/x/registry"
)

var recordFilePath = "../data/examples/service_provider_example.yml"

func init() {
	var err error
	if recordFilePath, err = filepath.Abs(recordFilePath); err != nil {
		panic(err)
	}
	if _, err = os.Stat(recordFilePath); err != nil {
		panic(err)
	}
}

type registrySuite struct {
	testSuite

	accountAddress string
	bondId         string
}

func TestRegistry(t *testing.T) {
	suite.Run(t, new(registrySuite))
}

func (s *registrySuite) SetupTest() {
	s.resetChain()

	Sut.ModifyGenesisJSON(s.T(), UpdateGenesisRegistry(s))
	s.accountAddress = s.cli().GetKeyAddr(s.account(0))

	Sut.StartChain(s.T())
	s.bondId = createBond(&s.testSuite)
}

func (s *registrySuite) TestTxSetRecord() {
	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"request with invalid payload file arg",
			[]string{"bad-file", s.bondId},
			true,
		},
		{
			"success",
			[]string{recordFilePath, s.bondId},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := []string{"tx", "registry", "set"}
			cmd = append(cmd, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *testSuite) updateParams(params *registry.Params) {
	params.RecordRent = sdk.NewCoin(s.bondDenom, math.NewInt(1000))
	params.RecordRentDuration = 10 * time.Second

	params.AuthorityRent = sdk.NewCoin(s.bondDenom, math.NewInt(1000))
	params.AuthorityGracePeriod = 10 * time.Second

	params.AuthorityAuctionCommitFee = sdk.NewCoin(s.bondDenom, math.NewInt(100))
	params.AuthorityAuctionRevealFee = sdk.NewCoin(s.bondDenom, math.NewInt(100))
	params.AuthorityAuctionMinimumBid = sdk.NewCoin(s.bondDenom, math.NewInt(500))
}

func UpdateGenesisRegistry(s *registrySuite) systest.GenesisMutator {
	require := s.Require()

	return func(genesis []byte) []byte {
		var regState registry.GenesisState
		raw := gjson.Get(string(genesis), "app_state.registry").String()
		require.NoError(s.codec.UnmarshalJSON([]byte(raw), &regState))

		s.updateParams(&regState.Params)
		regStateBz, err := s.codec.MarshalJSON(&regState)
		require.NoError(err)
		genesis, err = sjson.SetRawBytes(genesis, "app_state.registry", regStateBz)
		require.NoError(err)

		// // add consensus-controlled module account(s)
		// bondinitmsg, err := codectypes.NewAnyWithValue(&nitro.MsgInitAccount{})
		// require.NoError(s.T(), err)
		// var accountsState accounts.GenesisState
		// raw = gjson.Get(string(genesis), "app_state.accounts").String()
		// require.NoError(s.T(), cdc.UnmarshalJSON([]byte(raw), &accountsState))
		// accountsState.InitAccountMsgs = []*accounts.MsgInit{
		// 	{
		// 		Sender:      accountAddress,
		// 		AccountType: "nitro",
		// 		Message:     bondinitmsg,
		// 	},
		// }
		// accountsStateBz, err := cdc.MarshalJSON(&accountsState)
		// require.NoError(s.T(), err)
		// genesis, err = sjson.SetRawBytes(genesis, "app_state.accounts", accountsStateBz)
		// require.NoError(s.T(), err)

		return genesis
	}
}

func (s *registrySuite) reserveName(authorityName string) {
	cmd := []string{
		"tx", "registry", "reserve-authority", authorityName, s.accountAddress,
	}
	s.runTx(cmd, false)
}

func (s *registrySuite) createNameRecord(authorityName string) {
	for _, cmd := range [][]string{
		// reserve name authority
		{"tx", "registry", "reserve-authority", authorityName, s.accountAddress},
		// add bond-id to name authority
		{"tx", "registry", "authority-bond", authorityName, s.bondId},
		// create actual name record
		{"tx", "registry", "set-name", fmt.Sprintf("lrn://%s/", authorityName), "test_hello_cid"},
	} {
		s.runTx(cmd, false)
	}
}

func (s *registrySuite) createRecord(bondId string) {
	s.T().Helper()
	cmd := []string{
		"tx", "registry", "set", recordFilePath, bondId,
	}
	s.runTx(cmd, false)
}
