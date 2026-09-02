package nitro

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync/atomic"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/pflag"
	nitrocrypto "github.com/statechannels/go-nitro/crypto"
	"github.com/statechannels/go-nitro/node"
	"github.com/statechannels/go-nitro/node/engine"
	"github.com/statechannels/go-nitro/node/engine/chainservice"
	chainutils "github.com/statechannels/go-nitro/node/engine/chainservice/utils"
	"github.com/statechannels/go-nitro/node/engine/store"
	nitrotypes "github.com/statechannels/go-nitro/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/utils"
)

const (
	serverName = "nitro"

	ServerContextKey = "server." + serverName
)

// TODO
// 1. implement participation as a signing group member:
//    1.1. initialize group delegate credentials, with distsig over ABCI
//    1.2. bind custodian contract for chain/adjudicator actions
// 2. operation to update DKG group members
// 3. integrate with onboarding module

// Server wraps and configures a go-nitro Node.
type Server struct {
	logger       log.Logger
	config       *Config
	node         *node.Node
	storeDir     string // path to Nitro store directory
	reactor      *p2pReactor
	keyring      keyring.Keyring
	addressCodec address.Codec
	ethAddress   nitrotypes.Address
	nodeKey      cmtcrypto.PrivKey
	// participantID nitrotypes.Participant
	ready atomic.Bool
}

func New(globalConfig server.ConfigMap, logger log.Logger, kr keyring.Keyring) (*Server, error) {
	home, _ := globalConfig[flags.FlagHome].(string)
	cfg, err := UnmarshalConfig(globalConfig)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s := &Server{
		config:       cfg,
		storeDir:     filepath.Join(home, "nitro"),
		reactor:      newReactor(cfg, logger.With(log.ModuleKey, serverName, "component", "p2p-reactor")),
		keyring:      kr,
		logger:       logger.With(log.ModuleKey, serverName),
		addressCodec: utils.NewAddressCodec(),
	}
	return s, nil
}

func (*Server) Name() string {
	return serverName
}

func (s *Server) Start(ctx context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}
	// TODO: pass keyring through context?
	if err := s.init(s.keyring); err != nil {
		return err
	}
	s.logger.Info("starting Nitro server")
	s.ready.Store(true)
	return nil
}

func (s *Server) Stop(context.Context) error {
	s.logger.Info("stopping Nitro server")
	s.ready.Store(false)
	if s.node != nil {
		return s.node.Close()
	}
	return nil
}

func (s *Server) init(kr keyring.Keyring) error {
	var (
		ethkey cmtcrypto.PrivKey
		err    error
		c      = s.config
	)
	if c.EthKey != "" {
		ethkey, err = utils.ExtractPrivateKeyByUid(kr, c.EthKey)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("eth-key is not set") // should already be caught by Config.Validate
	}
	s.nodeKey = ethkey

	storeOpts := store.StoreOpts{
		UseDurableStore: true,
		DurableStoreDir: s.storeDir,
	}
	chainOpts := chainservice.ChainOpts{
		ChainUrl:       c.EthURL,
		StartBlockNum:  c.EthStartBlock,
		ChainAuthToken: c.EthAuthToken,
		ChainPk:        hex.EncodeToString(ethkey.Bytes()),
	}
	// inject SDK logger into slog, which Nitro uses
	loggerImpl, ok := s.logger.Impl().(*slog.Logger)
	if ok {
		slog.SetDefault(loggerImpl)
	} else {
		s.logger.Warn("Logger backend is not slog, cannot set as Nitro logger")
	}

	store, err := store.NewStore(storeOpts)
	if err != nil {
		return err
	}
	// Compare chainOpts.ChainStartBlock to lastBlockNum seen in store. The larger of the two
	// gets passed as an argument when creating NewEthChainService
	storeBlockNum, err := store.GetLastBlockNumSeen()
	if err != nil {
		return err
	}
	chainOpts.StartBlockNum = max(storeBlockNum, chainOpts.StartBlockNum)

	contractAddresses := node.ContractAddresses{
		NaAddress:  common.HexToAddress(c.EthNaAddress),
		VpaAddress: common.HexToAddress(c.EthVpaAddress),
		CaAddress:  common.HexToAddress(c.EthCaAddress),
	}
	chainOpts.CreateAdjudicator = chainutils.AdjudicatorCreator(contractAddresses.NaAddress)
	chain, err := chainservice.NewL1ChainService(chainOpts)
	if err != nil {
		return err
	}
	// note: utils.EthAddressFromPubKey expects uncompressed keys
	ethPubkey, err := crypto.DecompressPubkey(ethkey.PubKey().Bytes())
	if err != nil {
		return err
	}
	s.ethAddress = nitrotypes.Address(crypto.PubkeyToAddress(*ethPubkey))

	s.node = node.New(
		newMessageService("laconicd-p2p-"+s.ethAddress.String(), s.reactor),
		chain,
		store,
		&engine.PermissivePolicy{},
		nitrocrypto.NewSimpleCredential(ethkey.Bytes()),
		contractAddresses,
	)
	return nil
}

func (s *Server) Config() any {
	if s.config == nil {
		return DefaultConfig()
	}
	return s.config
}

func (s *Server) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(s.Name(), pflag.ExitOnError)
	AddFlags(flags)
	return flags
}

func (s *Server) Ready() bool {
	return s.ready.Load()
}

func (s *Server) ContextKey() string {
	return ServerContextKey
}

func (s *Server) P2PReactors() map[string]p2p.Reactor {
	return map[string]p2p.Reactor{"NITRO": s.reactor}
}

func (s *Server) nodeAddress() sdk.AccAddress {
	return sdk.AccAddress(s.nodeKey.PubKey().Address())
}
