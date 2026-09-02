package distsig

import (
	"context"
	"fmt"
	"math"
	"math/big"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/ethereum/go-ethereum/common"
	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	clientdss "git.vdb.to/cerc-io/chain-signatures/ethdss"
	"git.vdb.to/cerc-io/chain-signatures/ethschnorr"
	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/utils"
)

type (
	DealMap = map[int]*dkg.Deal

	DkgRunID int64
	SigRunID string
)

const (
	componentName = "distsig"
)

type Manager struct {
	config  *Config
	logger  log.Logger
	keyring keyring.Keyring

	longtermKey    Scalar // only initialized after Start()
	longtermPubKey cmtcrypto.PubKey

	dkgs         map[DkgRunID]*dkgRun
	currentDkgID DkgRunID // zero indicates no current run
	sigs         map[SigRunID]*sigRun
	completeSigs []SigRunID
}

type PeerMessages struct {
	dkg *DkgMessages
	dss []SigMessages
}

func New(logger log.Logger, globalConfig server.ConfigMap, kr keyring.Keyring) (*Manager, error) {
	config, err := UnmarshalConfig(globalConfig)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		config:  config,
		keyring: kr,
		dkgs:    make(map[DkgRunID]*dkgRun),
		sigs:    make(map[SigRunID]*sigRun),
	}
	m.logger = logger.With(log.ModuleKey, m.Name())
	return m, nil
}

func (*Manager) Name() string { return componentName }

func (m *Manager) Start(ctx context.Context) error {
	if !m.config.Enable {
		m.logger.Info(fmt.Sprintf("%s server is disabled via config", m.Name()))
		return nil
	}
	if m.config.LongtermKey == "" {
		return fmt.Errorf("missing longterm key")
	}
	longtermPrivKey, err := utils.ExtractPrivateKeyByUid(m.keyring, m.config.LongtermKey)
	if err != nil {
		return fmt.Errorf("failed to extract longterm key: %w", err)
	}
	m.longtermKey = suite.Scalar().SetBytes(longtermPrivKey.Bytes())
	m.longtermPubKey = longtermPrivKey.PubKey()
	return nil
}

func (*Manager) Stop(context.Context) error { return nil }

func (m *Manager) Config() any {
	if m.config == nil {
		return DefaultConfig()
	}
	return m.config
}

func (m *Manager) getDkg(runid DkgRunID) (*dkgRun, error) {
	if runid == 0 {
		// runid = m.currentDkg
		return nil, fmt.Errorf("invalid run ID")
	}
	if run, ok := m.dkgs[runid]; ok {
		return run, nil
	}
	return nil, fmt.Errorf("DKG not initialized for run: %v", runid)
}

func (m *Manager) currentDkg() *dkgRun {
	run, err := m.getDkg(m.currentDkgID)
	if err != nil {
		panic(err)
	}
	return run
}

func (m *Manager) LongtermPublicKey() cmtcrypto.PubKey {
	return m.longtermPubKey
}

func (m *Manager) LongtermEthAddress() common.Address {
	pubkey, err := SuitePublicKeyFromBytes(m.LongtermPublicKey().Bytes())
	if err != nil {
		panic(fmt.Errorf("failed to parse longterm pubkey: %w", err))
	}
	return pubkey.Address()
}

func thresholdForRatio(m int, tr float64) int {
	return int(math.Ceil(float64(m) * tr))
}

func (m *Manager) initDKG(members []Point) (*dkgRun, error) {
	t := thresholdForRatio(len(members), m.config.ThresholdRatio)
	keygen, err := dkg.NewDistKeyGenerator(suite.(dkg.Suite), m.longtermKey, members, t)
	if err != nil {
		return nil, err
	}
	return &dkgRun{DistKeyGenerator: keygen}, nil
}

// NeedDKG returns true if DKG needs to be run for the given participants.
func (m *Manager) NeedDKG(pubkeys []Point) bool {
	return m.currentDkgID == 0 || !equalPoints(m.currentDkg().Participants(), pubkeys)
}

// StartDKG begins a new DKG run including the given participants.
func (m *Manager) StartDKG(block int64, pubkeys []Point) error {
	// TODO: generate runid from block height and hash of pubkeys?
	runid := DkgRunID(block)
	if _, ok := m.dkgs[runid]; ok {
		return fmt.Errorf("DKG run %v already exists", runid)
	}
	if len(pubkeys) < 2 {
		m.currentDkgID = 0
		m.logger.Debug("Too few participants for distributed signature")
		return nil
	}
	run, err := m.initDKG(pubkeys)
	if err != nil {
		return err
	}
	m.dkgs[runid] = run
	m.currentDkgID = runid
	return run.prepareMessages()
}

func (m *Manager) initDSS(dkgid DkgRunID, msg *big.Int) (*sigRun, error) {
	run, err := m.getDkg(dkgid)
	if err != nil {
		return nil, err
	}
	if !run.Finished() {
		return nil, fmt.Errorf("DKG run has not finished: %v", dkgid)
	}
	random, err := run.DistKeyShare()
	if err != nil {
		return nil, err
	}
	dss, err := clientdss.NewDSS(clientdss.DSSArgs{
		Secret:       m.longtermKey,
		Participants: run.Participants(),
		Long:         run.share,
		Random:       random,
		Msg:          msg,
		T:            run.Threshold(),
		// Qualified:    run.QUAL(),
	})
	if err != nil {
		return nil, err
	}
	return &sigRun{DSS: dss, dkgID: dkgid}, nil
}

// StartSignature begins a DSS run for the given message and DKG state.
func (m *Manager) StartSignature(msg *big.Int) error {
	if m.currentDkgID == 0 {
		return fmt.Errorf("no distributed key prepared")
	}
	run, err := m.initDSS(m.currentDkgID, msg)
	if err != nil {
		return err
	}
	sigid := run.SigRunID()
	m.sigs[sigid] = run
	return run.prepareMessages()
}

func (m *Manager) CompletedSignatures() map[SigRunID]ethschnorr.Signature {
	var ret map[SigRunID]ethschnorr.Signature
	for _, id := range m.completeSigs {
		run, ok := m.sigs[id]
		if !ok {
			panic(fmt.Errorf("DSS run not found: %v", id))
		}
		if run.sig == nil {
			panic(fmt.Errorf("DSS signature not completed: %v", id))
		}
		// Deterministically pick a participant responsible for submitting the signature
		session := new(big.Int).SetBytes(run.SessionID()[:8])
		submitterIdx := session.Uint64() % uint64(len(run.Participants()))
		if run.Index() == int(submitterIdx) {
			ret[id] = run.sig
		}
	}
	return ret
}

// FlushMessages flushes the message buffers for any active DKG and DSS runs.
func (m *Manager) FlushMessages() PeerMessages {
	var ret PeerMessages
	if buf := m.currentDkg().flushMessages(); len(buf) != 0 {
		ret.dkg = &DkgMessages{runID: m.currentDkgID, messages: buf}
	}
	for id, run := range m.sigs {
		if buf := run.flushMessages(); buf != nil {
			ret.dss = append(ret.dss, SigMessages{id, buf})
		}
	}
	return ret
}

// ProcessMessages processes DKG and DSS peer messages.
func (m *Manager) ProcessMessages(dm *PeerMessages) error {
	if dkg := dm.dkg; dkg != nil {
		run, err := m.getDkg(dkg.runID)
		if err != nil {
			return err
		}
		if err := run.processMessages(dkg.messages); err != nil {
			return err
		}
	}
	for _, dss := range dm.dss {
		run, ok := m.sigs[dss.runID]
		if !ok {
			return fmt.Errorf("DSS not initialized for run: %v", dss.runID)
		}
		if err := run.processMessage(dss.messages); err != nil {
			return err
		}
		if run.sig != nil {
			m.completeSigs = append(m.completeSigs, dss.runID)
		}
	}
	return nil
}

func (pm PeerMessages) Empty() bool {
	return pm.dkg == nil && len(pm.dss) == 0
}
