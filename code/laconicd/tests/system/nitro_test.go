package system

import (
	"encoding/json"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"testing"

	nitrocrypto "github.com/statechannels/go-nitro/crypto"
	chainutils "github.com/statechannels/go-nitro/node/chain"
	"github.com/statechannels/go-nitro/node/query"
	"github.com/statechannels/go-nitro/node_test"
	nitrotypes "github.com/statechannels/go-nitro/types"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

var (
	ethChainPort = "8546"
	ethChainUrl  = "ws://127.0.0.1:8546"
)

type nitroSuite struct {
	testSuite

	tokenDenom   string
	ethChainOpts *node_test.LocalChainOptions

	alice, bob     nitrocrypto.Credential
	intermediaries []nitrocrypto.Credential
}

func TestNitro(t *testing.T) {
	suite.Run(t, new(nitroSuite))
}

func (s *nitroSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	s.tokenDenom = "eth"

	// Set up Ethereum test chain
	// TODO: use anvil for better self-containment
	ethChainOpts := node_test.ExternalHardhatChainOptions(s.T())
	ethChainOpts.ChainUrl = ethChainUrl
	aliceKey := ethChainOpts.ChainPks[0]
	bobKey := ethChainOpts.ChainPks[1]
	s.ethChainOpts = ethChainOpts

	ethChain, err := node_test.ConnectLocalChain(ethChainOpts, nitrotypes.BytesFromHex(aliceKey))
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(ethChain.Close()) })

	s.alice = nitrocrypto.NewSimpleCredential(nitrotypes.BytesFromHex(aliceKey))
	s.bob = nitrocrypto.NewSimpleCredential(nitrotypes.BytesFromHex(bobKey))
	for _, pk := range ethChainOpts.ChainPks {
		creds := nitrocrypto.NewSimpleCredential(nitrotypes.BytesFromHex(pk))
		s.intermediaries = append(s.intermediaries, creds)
	}

	s.configureEthChain()
	// TODO when using intermediaries, set more validator node pks here
	s.setValidatorNitroKeys(s.ethChainOpts.ChainPks[1])
}

func (s *nitroSuite) SetupTest() {
	s.resetChain()

	s.setClientKeys(s.ethChainOpts.ChainPks[0])

	Sut.StartChain(s.T())
}

func (s *nitroSuite) SetupSubTest() { s.SetupTest() }

func (s *nitroSuite) configureEthChain() {
	cas, err := chainutils.LoadEnvContractAddresses()
	s.Require().NoError(err)

	cmds := [][]string{
		{"config", "set", "app", "nitro.eth-na-address", "'" + cas.NaAddress.String() + "'"},
		{"config", "set", "app", "nitro.eth-vpa-address", "'" + cas.VpaAddress.String() + "'"},
		{"config", "set", "app", "nitro.eth-ca-address", "'" + cas.CaAddress.String() + "'"},
		{"config", "set", "app", "nitro.eth-url", ethChainUrl},
	}
	_ = Sut.ForEachNodeExecAndWait(s.T(), cmds...)
	for _, cmd := range cmds {
		cmd = append(cmd, "--home", s.clientHome())
		_ = systest.MustRunShellCmd(s.T(), Sut.ExecBinary(), cmd...)
	}
}

func (s *nitroSuite) TestTxOpenChannel() {
	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"missing args",
			[]string{},
			true,
		},
		{
			"missing amount",
			[]string{s.bob.AsParticipant().String()},
			true,
		},
		{
			"valid request",
			[]string{s.bob.AsParticipant().String(), "10" + s.tokenDenom},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := append([]string{"tx", "nitro", "open-channel"}, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *nitroSuite) TestTxCloseChannel() {
	testCases := []struct {
		name  string
		args  []string
		err   bool
		setup func() []string
	}{
		{
			"missing args",
			[]string{},
			true,
			nil,
		},
		{
			"invalid channel",
			[]string{"invalid-channel-id"},
			true,
			nil,
		},
		{
			"valid channel",
			nil,
			false,
			func() []string { return s.setupLedgerChannel("10") },
		},
		// {
		//	"valid channel with challenge",
		//	[]string{"--challenge"},
		//	false,
		//	func() []string { return s.setupLedgerChannel("10") },
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := []string{"tx", "nitro", "close-channel"}
			if tc.setup != nil {
				cmd = append(cmd, tc.setup()...)
			}
			cmd = append(cmd, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *nitroSuite) TestTxOpenPaymentChannel() {

	testCases := []struct {
		name  string
		args  []string
		err   bool
		setup func() []string
	}{
		{
			"missing args",
			[]string{},
			true,
			nil,
		},
		{
			"missing amount",
			[]string{s.bob.AsParticipant().String()},
			true,
			nil,
		},
		{
			"zero hops",
			[]string{s.bob.AsParticipant().String(), "10" + s.tokenDenom},
			false,
			func() []string { s.setupLedgerChannel("20"); return nil },
		},
		// { // TODO open n-hop ledger channels
		//	"one hop",
		//	[]string{"10" + s.tokenDenom, s.bob.AsParticipant().String(), s.intermediaries[0].String()},
		//	false,
		//	setupLedgerChannelForPaymentChannel,
		// },
		// { // TODO use a different counterparty, cannot have duplicate channels
		//	"direct channel with challenge duration",
		//	[]string{"10" + s.tokenDenom, s.bob.AsParticipant().String(), "--challenge-duration", "100"},
		//	false,
		//	setupLedgerChannelForPaymentChannel,
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := []string{"tx", "nitro", "open-payment-channel"}
			if tc.setup != nil {
				tc.setup()
			}
			cmd = append(cmd, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *nitroSuite) TestTxClosePaymentChannel() {

	testCases := []struct {
		name  string
		args  []string
		err   bool
		setup func() []string
	}{
		{
			"missing args",
			[]string{},
			true,
			nil,
		},
		// { // TODO use different counterparty
		//	"invalid channel ID",
		//	[]string{"invalid-channel-id"},
		//	true,
		//	setupLedgerChannel,
		// },
		{
			"zero hops",
			nil,
			false,
			func() []string {
				s.setupLedgerChannel("20")
				return s.setupPaymentChannel("10")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := []string{"tx", "nitro", "close-payment-channel"}
			if tc.setup != nil {
				cmd = append(cmd, tc.setup()...)
			}
			cmd = append(cmd, tc.args...)
			s.runTx(cmd, tc.err)
		})
	}
}

func (s *nitroSuite) TestTxPay() {

	testCases := []struct {
		name  string
		args  []string
		error bool
		setup func() []string
		paid  *big.Int
	}{
		{
			name:  "missing args",
			error: true,
		},
		{
			name:  "invalid channel ID",
			args:  []string{"invalid-channel-id", "10" + s.tokenDenom},
			error: true,
		},
		{
			name:  "invalid amount format",
			args:  []string{"0x0102030000000000000000000000000000000000000000000000000000000000", "invalid-amount"},
			error: true,
		},
		{
			name:  "nonexistent channel",
			args:  []string{"0x0102030000000000000000000000000000000000000000000000000000000000", "5" + s.tokenDenom},
			error: true,
		},
		{
			name: "valid payment",
			args: []string{"10" + s.tokenDenom},
			setup: func() []string {
				s.setupLedgerChannel("100")
				return s.setupPaymentChannel("50")
			},
			paid: big.NewInt(10),
		},
		// {
		//	"insufficient funds",
		//	[]string{"10" + s.tokenDenom},
		//	false,
		//	func() []string {
		//		s.setupLedgerChannel("100")
		//		return s.setupPaymentChannel("50")
		//	},
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var channelId string

			cmd := []string{"tx", "nitro", "pay"}
			if tc.setup != nil {
				setupArgs := tc.setup()
				cmd = append(cmd, setupArgs...)

				if !tc.error && len(setupArgs) > 0 {
					channelId = setupArgs[0]
				}
			}
			cmd = append(cmd, tc.args...)
			s.runTx(cmd, tc.error)

			// Check balance change for successful payments
			if channelId != "" {
				balance := s.getChannelBalance(channelId)
				s.Require().Equal(tc.paid, balance.PaidSoFar.ToInt())
			}
		})
	}
}

func (s *nitroSuite) TestFullPaymentFlow() {
	s.Run("complete payment lifecycle", func() {
		// Open ledger and payment channels
		ledgerChannel := s.setupLedgerChannel("100")[0]
		paymentChannel := s.setupPaymentChannel("50")[0]

		// Verify initial balance
		s.Require().Equal("0", s.getChannelBalance(paymentChannel).PaidSoFar.ToInt().String())

		// Make payments: 5 + 10 + 15 = 30
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "5" + s.tokenDenom}, false)
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "10" + s.tokenDenom}, false)
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "15" + s.tokenDenom}, false)

		// Verify final balance
		finalBalance := s.getChannelBalance(paymentChannel)
		s.Require().Equal("30", finalBalance.PaidSoFar.ToInt().String())
		s.Require().Equal("20", finalBalance.RemainingFunds.ToInt().String())

		// Close channels in correct order
		s.runTx([]string{"tx", "nitro", "close-payment-channel", paymentChannel}, false)
		s.runTx([]string{"tx", "nitro", "close-channel", ledgerChannel}, false)
	})

	s.Run("insufficient funds scenario", func() {
		s.setupLedgerChannel("20")
		paymentChannel := s.setupPaymentChannel("10")[0]

		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "5" + s.tokenDenom}, false)
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "10" + s.tokenDenom}, true) // insufficient funds
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "5" + s.tokenDenom}, false)

		// Verify all funds used
		balance := s.getChannelBalance(paymentChannel)
		s.Require().Equal("10", balance.PaidSoFar.ToInt().String())
		s.Require().Equal("0", balance.RemainingFunds.ToInt().String())
	})

	s.Run("invalid closure order", func() {
		// Open both channels
		ledgerChannel := s.setupLedgerChannel("50")[0]
		paymentChannel := s.setupPaymentChannel("30")[0]

		// Make a payment
		s.runTx([]string{"tx", "nitro", "pay", paymentChannel, "10" + s.tokenDenom}, false)

		// Try to close ledger channel before payment channel - should fail
		s.runTx([]string{"tx", "nitro", "close-channel", ledgerChannel}, true)

		// Correct order should work
		s.runTx([]string{"tx", "nitro", "close-payment-channel", paymentChannel}, false)
		s.runTx([]string{"tx", "nitro", "close-channel", ledgerChannel}, false)
	})
}

// setupLedgerChannel creates a ledger channel and returns the channel ID as args
func (s *nitroSuite) setupLedgerChannel(amount string) []string {
	openCmd := []string{"tx", "nitro", "open-channel", s.bob.AsParticipant().String(), amount + s.tokenDenom}
	txhash := s.runTx(openCmd, false)
	return []string{s.getChannelIdFromTx(txhash)}
}

// setupPaymentChannel creates a ledger channel and payment channel, returns payment channel ID as args
func (s *nitroSuite) setupPaymentChannel(amount string) []string {
	openPcCmd := []string{"tx", "nitro", "open-payment-channel", s.bob.AsParticipant().String(), amount + s.tokenDenom}
	txhash := s.runTx(openPcCmd, false)
	return []string{s.getChannelIdFromTx(txhash)}
}

// getChannelIdFromTx opens a channel and extracts the channel ID from the transaction result
func (s *nitroSuite) getChannelIdFromTx(txHash string) string {
	txDetails := s.cli().CustomQuery("query", "tx", txHash)

	// find channel_id in any event's attributes
	channelId := gjson.Get(txDetails, "events.#.attributes.#(key==\"channel_id\").value|0")
	s.Require().True(channelId.Exists(), "channel_id not found in transaction events")

	return channelId.String()
}

func (s *nitroSuite) getChannelBalance(channelId string) query.PaymentChannelBalance {
	channelInfo := s.cli().CustomQuery(
		s.withNitroQueryFlags("q", "nitro", "get-channel", "--payment", channelId)...,
	)

	var info query.PaymentChannelInfo
	err := json.Unmarshal([]byte(channelInfo), &info)
	s.Require().NoError(err, "could not parse PaymentChannelInfo: %s", channelInfo)
	return info.Balance
}

func (s *nitroSuite) withNitroQueryFlags(cmd ...string) []string {
	return append(cmd, "--home", s.clientHome())
}

func (s *nitroSuite) Test_Hang() {
	hang := os.Getenv("HANG")
	if hang == "" {
		s.T().Skip("Placeholder to run the validator nodes and hang for manual testing")
	}

	if hang == "ledger" || hang == "payment" {
		s.setupLedgerChannel("20")
	}
	if hang == "payment" {
		s.setupPaymentChannel("10")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
