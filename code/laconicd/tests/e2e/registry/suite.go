package registry

import (
	"fmt"
	"path/filepath"
	"time"

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
	bondcli "git.vdb.to/cerc-io/laconicd/x/bond/client/cli"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/client/cli"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	accountName    string
	accountAddress string

	bondId string
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (ets *E2ETestSuite) SetupSuite() {
	sr := ets.Require()
	ets.T().Log("setting up e2e test suite")

	var err error

	genesisState := ets.cfg.GenesisState
	var registryGenesis registrytypes.GenesisState
	ets.Require().NoError(ets.cfg.Codec.UnmarshalJSON(genesisState[registrytypes.ModuleName], &registryGenesis))

	ets.updateParams(&registryGenesis.Params)

	registryGenesisBz, err := ets.cfg.Codec.MarshalJSON(&registryGenesis)
	ets.Require().NoError(err)
	genesisState[registrytypes.ModuleName] = registryGenesisBz
	ets.cfg.GenesisState = genesisState

	ets.network, err = network.New(ets.T(), ets.T().TempDir(), ets.cfg)
	sr.NoError(err)

	_, err = ets.network.WaitForHeight(2)
	sr.NoError(err)

	// setting up random account
	ets.accountName = "accountName"
	ets.createAccountWithBalance(ets.accountName, &ets.accountAddress)

	ets.bondId = ets.createBond()
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
		sdk.NewCoins(sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(100000000))),
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
	createBondCmd := bondcli.NewCreateBondCmd()
	args := []string{
		fmt.Sprintf("1000000%s", ets.cfg.BondDenom),
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
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))

	// getting the bonds list and returning the bond-id
	clientCtx := val.ClientCtx
	cmd := bondcli.GetQueryBondList()
	args = []string{
		fmt.Sprintf("--%s=json", flags.FlagOutput),
	}
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	var queryResponse bondtypes.QueryBondsResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &queryResponse)
	sr.NoError(err)

	// extract bond id from bonds list
	bond := queryResponse.GetBonds()[0]
	return bond.GetId()
}

func (ets *E2ETestSuite) reserveName(authorityName string) {
	val := ets.network.Validators[0]
	sr := ets.Require()

	clientCtx := val.ClientCtx
	cmd := cli.GetCmdReserveAuthority()
	args := []string{
		authorityName,
		ets.accountAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)

	var d sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))
}

func (ets *E2ETestSuite) createNameRecord(authorityName string) {
	val := ets.network.Validators[0]
	sr := ets.Require()

	// reserving the name
	clientCtx := val.ClientCtx
	cmd := cli.GetCmdReserveAuthority()
	args := []string{
		authorityName,
		ets.accountAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	var d sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))

	// Get the bond-id
	bondId := ets.bondId

	// adding bond-id to name authority
	args = []string{
		authorityName, bondId,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	cmd = cli.GetCmdSetAuthorityBond()

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))

	args = []string{
		fmt.Sprintf("lrn://%s/", authorityName),
		"test_hello_cid",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}

	cmd = cli.GetCmdSetName()

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))
}

func (ets *E2ETestSuite) createRecord(bondId string) {
	val := ets.network.Validators[0]
	sr := ets.Require()

	payloadPath := "../../data/examples/service_provider_example.yml"
	payloadFilePath, err := filepath.Abs(payloadPath)
	sr.NoError(err)

	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, ets.accountName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=json", flags.FlagOutput),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fmt.Sprintf("3%s", ets.cfg.BondDenom)),
	}
	args = append([]string{payloadFilePath, bondId}, args...)
	clientCtx := val.ClientCtx
	cmd := cli.GetCmdSetRecord()

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	sr.NoError(err)
	var d sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &d)
	sr.NoError(err)
	sr.NoError(laconictestcli.CheckTxCode(ets.network, val.ClientCtx, d.TxHash, 0))
}

func (ets *E2ETestSuite) updateParams(params *registrytypes.Params) {
	params.RecordRent = sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(1000))
	params.RecordRentDuration = 10 * time.Second

	params.AuthorityRent = sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(1000))
	params.AuthorityGracePeriod = 10 * time.Second

	params.AuthorityAuctionCommitFee = sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(100))
	params.AuthorityAuctionRevealFee = sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(100))
	params.AuthorityAuctionMinimumBid = sdk.NewCoin(ets.cfg.BondDenom, math.NewInt(500))
}
