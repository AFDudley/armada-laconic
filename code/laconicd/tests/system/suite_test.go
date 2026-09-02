package system

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/app/params"
)

type testSuite struct {
	suite.Suite

	codec             codec.Codec
	accountNamePrefix string
	fees              string
	bondDenom         string

	clientEthKey    string
	validatorEthKey string
}

func (s *testSuite) SetupSuite() {
	s.codec = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	s.accountNamePrefix = "node" // from testnet (--node-dir-prefix)
	s.fees = "1" + params.CoinUnit
	s.bondDenom = sdk.DefaultBondDenom

	s.validatorEthKey = "validator-eth-key"
	s.clientEthKey = "client-eth-key"

	// configure the validators and client
	commonInitCommands := [][]string{
		{"config", "set", "client", "chain-id", "testing"},
		{"config", "set", "client", "keyring-backend", "test"},
	}

	// disable nitro on validators by default
	Sut.ForEachNodeExecAndWait(s.T(), append(commonInitCommands,
		[]string{"config", "set", "app", "nitro.enable", "false"},
	)...)

	clientInitCommands := append(commonInitCommands,
		[]string{"config", "set", "app", "nitro.eth-key", s.clientEthKey},
	)
	for _, cmd := range clientInitCommands {
		cmd = append(cmd, "--home", s.clientHome())
		_ = systest.MustRunShellCmd(s.T(), Sut.ExecBinary(), cmd...)
	}
}

func (s *testSuite) setValidatorNitroKeys(keys ...string) {
	Sut.WithEachNodeHome(func(i int, home string) {
		var cmds [][]string
		// TODO: complete and enable distsig
		// []string{"config", "set", "app", "distsig.enable", "true"},
		// []string{"config", "set", "app", "distsig.longterm-key", s.validatorNitroKey},
		// []string{"config", "set", "app", "nitro.use-distsig", "true"},

		if i < len(keys) {
			// configure validator as a payment channel (individual) counterparty
			cmds = [][]string{
				{"config", "set", "app", "nitro.enable", "true"},

				{"keys", "import-hex", s.validatorEthKey, keys[i]},
				{"config", "set", "app", "nitro.eth-key", s.validatorEthKey},
			}
		}

		for _, cmd := range cmds {
			systest.MustRunShellCmd(s.T(), Sut.ExecBinary(), append(cmd, "--home", home)...)
		}
	})
}

// note: keys have to be created after ResetChain, which clears the keyring
func (s *testSuite) setClientKeys(ethKey string) {
	commands := [][]string{
		{"keys", "import-hex", s.clientEthKey, ethKey},
	}
	for _, cmd := range commands {
		cmd = append(cmd, "--home", s.clientHome())
		_ = systest.MustRunShellCmd(s.T(), Sut.ExecBinary(), cmd...)
	}
}

func (s *testSuite) resetChain() {
	Sut.ResetChain(s.T())
	// clear the nitro data dirs
	Sut.WithEachNodeHome(func(i int, home string) {
		_ = os.RemoveAll(filepath.Join(home, "nitro"))

	})
	_ = os.RemoveAll(filepath.Join(s.clientHome(), "nitro"))
}

func (s *testSuite) account(i int) string {
	return fmt.Sprintf("%s%d", s.accountNamePrefix, i)
}

// the home dir systemtests uses to run client commands
func (s *testSuite) clientHome() string {
	return filepath.Join(workDir, "testnet")
}

func (s *testSuite) endpointURL(endpoint string) string {
	return fmt.Sprintf("%s/cerc/%s", Sut.APIAddress(), endpoint)
}

func (s *testSuite) cli() *systest.CLIWrapper {
	cli := systest.NewCLIWrapper(s.T(), Sut, verbose).
		WithRunStderr(os.Stderr)
	return &cli
}

func (s *testSuite) assertWithStderr(asserter systest.RunErrorAssert) systest.RunErrorAssert {
	return func(t assert.TestingT, err error, msgAndArgs ...any) bool {
		if asserter(t, err, msgAndArgs...) {
			return true
		}
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if stderr := exitError.Stderr; len(stderr) > 0 {
				defer t.Errorf("stderr: %s", stderr)
			}
		}
		return false
	}
}

func (s *testSuite) runTx(cmd []string, err bool) string {
	s.T().Helper()
	cli := s.cli()
	asserter := assert.NoError
	if err {
		asserter = assert.Error
	}
	// don't use Run(), it will clobber our custom --home
	out := cli.
		WithRunErrorMatcher(s.assertWithStderr(asserter)).
		RunCommandWithArgs(s.withTxFlags(cmd)...)
	if err {
		return ""
	}
	txResult, committed := cli.AwaitTxCommitted(out, txTimeout)
	s.Require().True(committed)
	systest.RequireTxSuccess(s.T(), txResult)
	txHash := gjson.Get(out, "txhash")
	s.Require().True(txHash.Exists(), "txhash not found in output: %s", out)
	return txHash.String()
}

func (s *testSuite) withTxFlags(cmd []string) []string {
	cmd = s.cli().WithTXFlags(cmd...)
	return append(cmd, []string{
		"--from", s.account(0),
		"--fees", s.fees,
	}...)
}
