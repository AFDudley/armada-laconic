package nitro

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/statechannels/go-nitro/protocols"
	nitrotypes "github.com/statechannels/go-nitro/types"

	"git.vdb.to/cerc-io/laconicd/server/relay"
	"git.vdb.to/cerc-io/laconicd/utils"
)

const txTimeoutDuration = 5 * time.Second

func GetServerFromCmd(cmd *cobra.Command) (*Server, error) {
	s, err := utils.GetFromContext[Server](cmd.Context(), ServerContextKey)
	if err != nil {
		return nil, err
	}
	for !s.Ready() {
		runtime.Gosched()
	}
	return s, nil
}

func delayAfterUpdate() {
	// HACK: delay so remote server status is synced before we return
	// TODO: correct solution is passing messages with consistency guarantees,
	// i.e. using txs to pass nitro messages and integrating state into chain storage
	time.Sleep(500 * time.Millisecond)
}

// waitForObjective waits for an objective to complete or fail using select pattern
func waitForObjective[Info any](
	failedChan <-chan protocols.ObjectiveId,
	objective protocols.ObjectiveId,
	updateChan <-chan Info,
	statusCheck func(Info) bool,
) error {
	for {
		select {
		case update := <-updateChan:
			if statusCheck(update) {
				delayAfterUpdate()
				return nil
			}
		case failedObjective := <-failedChan:
			if failedObjective == objective {
				return fmt.Errorf("objective failed for unspecified reason: %s", objective)
			}
		}
	}
}

func submitNitroTx(
	clientCtx client.Context,
	cmd *cobra.Command,
	msg MsgWrapper,
	unordered bool,
) error {
	if msg.SignerField != nil {
		ac := utils.NewAddressCodec()
		var err error
		*msg.SignerField, err = ac.BytesToString(clientCtx.FromAddress)
		if err != nil {
			return err
		}
	}
	factory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
	if err != nil {
		return err
	}
	if unordered {
		deadline := time.Now().Add(txTimeoutDuration)
		factory = factory.WithUnordered(true).WithTimeoutTimestamp(deadline)
	}
	if err := tx.GenerateOrBroadcastTxWithFactory(clientCtx, factory, msg.Msg); err != nil {
		return err
	}
	return nil
}

// setupNitroCommand sets up common infrastructure for nitro commands
func setupNitroCommand(cmd *cobra.Command) (*client.Context, *Server, func() error, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	s, err := GetServerFromCmd(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	relayServer, err := relay.GetServerFromCmd(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	sw, err := connectAsPeer(clientCtx, relayServer, s.P2PReactors())
	if err != nil {
		return nil, nil, nil, err
	}

	return &clientCtx, s, sw.Stop, nil
}

// setupNitroQueryCommand sets up infrastructure for nitro query commands (no tx context needed)
func setupNitroQueryCommand(cmd *cobra.Command, needp2p bool) (*Server, func() error, error) {
	s, err := GetServerFromCmd(cmd)
	if err != nil {
		return nil, nil, err
	}

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return nil, nil, err
	}

	if !needp2p {
		return s, func() error { return nil }, nil
	}

	relayServer, err := relay.GetServerFromCmd(cmd)
	if err != nil {
		return nil, nil, err
	}

	sw, err := connectAsPeer(clientCtx, relayServer, s.P2PReactors())
	if err != nil {
		return nil, nil, err
	}

	return s, sw.Stop, nil
}

// resolveTokens resolves amount and asset address to Nitro funds
func resolveTokens(s *Server, amountStr, assetAddr string) (nitrotypes.Funds, error) {
	var amount *big.Int
	var asset common.Address

	if assetAddr != "" {
		// Parse asset address
		if !common.IsHexAddress(assetAddr) {
			return nitrotypes.Funds{}, fmt.Errorf("invalid asset address: %s", assetAddr)
		}
		asset = common.HexToAddress(assetAddr)

		// A token specified by address must have an integer amount
		amount = new(big.Int)
		if _, ok := amount.SetString(amountStr, 10); !ok {
			return nitrotypes.Funds{}, fmt.Errorf("invalid amount for token %s: %s", assetAddr, amountStr)
		}
		if amount.Sign() <= 0 {
			return nitrotypes.Funds{}, fmt.Errorf("amount must be positive for token %s: %s", assetAddr, amountStr)
		}
	} else {
		// Parse as (known) coins and convert
		coins, err := sdk.ParseCoinsNormalized(amountStr)
		if err != nil {
			return nitrotypes.Funds{}, fmt.Errorf("failed to parse coins: %w", err)
		}
		if len(coins) != 1 {
			return nitrotypes.Funds{}, fmt.Errorf("only one asset is supported")
		}

		// Convert denomination to token address
		asset, err = denomToToken(coins[0].Denom)
		if err != nil {
			return nitrotypes.Funds{}, err
		}
		amount = coins[0].Amount.BigInt()
	}

	return nitrotypes.Funds{asset: amount}, nil
}

// map denoms to token addresses
// TODO nitro should control this mapping, and resolution should happen in the module logic
func denomToToken(denom string) (common.Address, error) {
	_tokens := map[string]common.Address{
		"eth": {},
	}
	token, ok := _tokens[strings.ToLower(denom)]
	if !ok {
		return common.Address{}, fmt.Errorf("unknown token %s", denom)
	}
	return token, nil
}

// connects as a peer to the relay server, initializing the P2P switch
func connectAsPeer(
	clientCtx client.Context, relayServer *relay.Server, reactors map[string]p2p.Reactor,
) (*relay.Switch, error) {
	// initialize the switch using the target's node info
	status, err := clientCtx.Client.Status(clientCtx.CmdContext)
	if err != nil {
		return nil, err
	}
	nodeInfo := status.NodeInfo
	sw, err := relayServer.GetSwitch(clientCtx.ChainID, nodeInfo.ProtocolVersion.Block)
	if err != nil {
		return nil, fmt.Errorf("error creating switch: %w", err)
	}
	for name, reactor := range reactors {
		if err := sw.AddReactor(name, reactor); err != nil {
			return nil, fmt.Errorf("error adding reactor: %w", err)
		}
	}
	if err := sw.Start(); err != nil {
		return nil, fmt.Errorf("error starting switch: %w", err)
	}

	peerAddr := p2p.IDAddressString(nodeInfo.ID(), nodeInfo.ListenAddr)
	// s.logger.Debug("dialing peer", "peer", peerAddr)
	if err := sw.Dial(peerAddr); err != nil {
		return nil, fmt.Errorf("error dialing peer: %w", err)
	}
	return sw, nil
}
