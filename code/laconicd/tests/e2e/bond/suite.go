package bond

import (
	"fmt"

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
	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	"git.vdb.to/cerc-io/laconicd/x/bond/client/cli"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	accountName    string
	accountAddress string
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

	// setting up random account
	ets.accountName = "accountName"
	ets.createAccountWithBalance(ets.accountName, &ets.accountAddress)
}

func (ets *E2ETestSuite) TearDownSuite() {
	ets.T().Log("tearing down e2e test suite")
	ets.network.Cleanup()
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

func (ets *E2ETestSuite) createBond() string {
	val := ets.network.Validators[0]
	sr := ets.Require()
	createBondCmd := cli.NewCreateBondCmd()
	args := []string{
		fmt.Sprintf("10%s", ets.cfg.BondDenom),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, createBondCmd, args)
	sr.NoError(err)

	var d sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.Zero(d.Code)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))

	// getting the bonds list and returning the bond-id
	clientCtx := val.ClientCtx
	cmd := cli.GetQueryBondList()
	args = []string{
		fmt.Sprintf("--%s=json", flags.FlagOutput),
	}
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	var queryResponse bondtypes.QueryBondsResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &queryResponse)
	sr.NoError(err)

	// extract bond id from bonds list
	bonds := queryResponse.GetBonds()
	sr.NotEmpty(bonds)

	return queryResponse.GetBonds()[0].GetId()
}
