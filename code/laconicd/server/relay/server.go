package relay

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"cosmossdk.io/log"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/pex"
	"github.com/cometbft/cometbft/version"
	cmtversion "github.com/cometbft/cometbft/version"

	"git.vdb.to/cerc-io/laconicd/server"
)

const (
	serverName = "relay"

	ServerContextKey = "server." + serverName
)

// Server is a server component with CometBFT p2p capabilities
type Server struct {
	logger      log.Logger
	config      *Config
	cometConfig *cmtcfg.Config
	nodeKey     *p2p.NodeKey
	ready       atomic.Bool
}

type Switch struct {
	*p2p.Switch
	transport *p2p.MultiplexTransport
}

func New(cfg server.ConfigMap, logger log.Logger) (*Server, error) {
	c, err := UnmarshalConfig(cfg)
	if err != nil {
		return nil, err
	}
	cometConfig := cmtcfg.DefaultConfig()
	if err := server.UnmarshalSubConfig(cfg, "", &cometConfig); err != nil {
		return nil, err
	}
	return &Server{
		config:      c,
		cometConfig: cometConfig,
		logger:      logger.With(log.ModuleKey, serverName),
	}, nil
}

func (*Server) Name() string {
	return serverName
}

func (s *Server) Start(ctx context.Context) error {
	var err error
	s.nodeKey, err = p2p.LoadOrGenNodeKey(s.cometConfig.NodeKeyFile())
	if err != nil {
		return err
	}

	s.ready.Store(true)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.ready.Store(false)
	return nil
}

func (s *Server) Config() any {
	if s.config == nil {
		return DefaultConfig()
	}
	return s.config
}

func (s *Server) Ready() bool {
	return s.ready.Load()
}

func (s *Server) ContextKey() string {
	return ServerContextKey
}

func (s *Server) GetSwitch(chainID string, blockVersion uint64) (*Switch, error) {
	nodeInfo, err := makeNodeInfo(s.cometConfig, s.nodeKey, chainID, blockVersion)
	if err != nil {
		return nil, err
	}
	transport := createTransport(s.cometConfig, nodeInfo, s.nodeKey)
	sw := p2p.NewSwitch(s.cometConfig.P2P, transport, p2p.SwitchPeerFilters())
	sw.SetNodeInfo(nodeInfo)
	sw.SetNodeKey(s.nodeKey)
	return &Switch{sw, transport}, nil
}

func (s *Switch) AddReactor(name string, reactor p2p.Reactor) error {
	s.Switch.AddReactor(name, reactor)
	// register the new channels to the nodeInfo
	ni, ok := s.NodeInfo().(p2p.DefaultNodeInfo)
	if !ok {
		return errors.New("Node info is not of type DefaultNodeInfo")
	}
	for _, chDesc := range reactor.GetChannels() {
		if !ni.HasChannel(chDesc.ID) {
			ni.Channels = append(ni.Channels, chDesc.ID)
			s.transport.AddChannel(chDesc.ID)
		}
	}
	s.SetNodeInfo(ni)
	return nil
}

func (s *Switch) Dial(peerAddr string) error {
	addr, err := p2p.NewNetAddressString(peerAddr)
	if err != nil {
		return fmt.Errorf("error resolving net address: %w", err)
	}
	err = s.AddPersistentPeers([]string{addr.String()})
	if err != nil {
		return fmt.Errorf("error adding persistent peer: %w", err)
	}
	return s.DialPeerWithAddress(addr)
}

func makeNodeInfo(
	config *cmtcfg.Config,
	nodeKey *p2p.NodeKey,
	chainID string,
	blockVersion uint64,
) (p2p.DefaultNodeInfo, error) {
	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: p2p.NewProtocolVersion(
			cmtversion.P2PProtocol,
			blockVersion,
			0,
		),
		DefaultNodeID: nodeKey.ID(),
		Network:       chainID,
		Version:       version.TMCoreSemVer,
		Channels:      nil, // updated in AddReactor
		Moniker:       config.Moniker,
		Other: p2p.DefaultNodeInfoOther{
			RPCAddress: config.RPC.ListenAddress,
		},
	}

	if config.P2P.PexReactor {
		nodeInfo.Channels = append(nodeInfo.Channels, pex.PexChannel)
	}

	lAddr := config.P2P.ExternalAddress

	if lAddr == "" {
		lAddr = config.P2P.ListenAddress
	}

	nodeInfo.ListenAddr = lAddr

	// err := nodeInfo.Validate()
	return nodeInfo, nil
}

func createTransport(
	config *cmtcfg.Config,
	nodeInfo p2p.NodeInfo,
	nodeKey *p2p.NodeKey,
) *p2p.MultiplexTransport {
	var (
		mConnConfig = p2p.MConnConfig(config.P2P)
		transport   = p2p.NewMultiplexTransport(nodeInfo, *nodeKey, mConnConfig)
		connFilters = []p2p.ConnFilterFunc{}
	)

	if !config.P2P.AllowDuplicateIP {
		connFilters = append(connFilters, p2p.ConnDuplicateIPFilter())
	}

	p2p.MultiplexTransportConnFilters(connFilters...)(transport)

	// Limit the number of incoming connections.
	max := config.P2P.MaxNumInboundPeers
	p2p.MultiplexTransportMaxIncomingConnections(max)(transport)

	return transport
}

func removeProtocolIfDefined(addr string) string {
	if strings.Contains(addr, "://") {
		return strings.Split(addr, "://")[1]
	}
	return addr
}
