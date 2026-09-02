package distsig

import (
	"errors"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"git.vdb.to/cerc-io/chain-signatures/ethschnorr"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"

	"git.vdb.to/cerc-io/laconicd/testutil"
)

func NewTestManager(t *testing.T) (*Manager, Point) {
	logger := log.NewTestLogger(t)
	kr := testutil.NewKeyring()
	accounts := sdktestutil.CreateKeyringAccounts(t, kr, 1)
	cfg := DefaultConfig()
	cfg.LongtermKey = accounts[0].Name

	dsm, err := New(logger, cfg, kr)
	require.NoError(t, err)

	longterm, err := kr.Key(cfg.LongtermKey)
	require.NoError(t, err)
	longtermPoint, err := KeyRecordToPoint(longterm)
	require.NoError(t, err)

	return dsm, longtermPoint
}

type round struct {
	name   string
	verify func(*Manager) error
	tweak  func([]PeerMessages)
}

var (
	dkgRounds = []round{
		{
			name: "deal",
		},
		{
			name: "response",
			verify: func(m *Manager) error {
				if !m.currentDkg().Certified() {
					return errors.New("not certified")
				}
				return nil
			},
		},
		{
			name: "commit",
			verify: func(m *Manager) error {
				if !m.currentDkg().Finished() {
					return errors.New("run not finished")
				}
				if m.currentDkg().share == nil {
					return errors.New("dist key share is absent")
				}
				return nil
			},
		},
	}
	dssRounds = []round{
		{
			name: "sign",
			verify: func(m *Manager) error {
				dss := m.sigs[m.completeSigs[0]]
				if dss.sig == nil {
					return errors.New("signature is nil")
				}
				dkg, err := m.getDkg(dss.dkgID)
				if err != nil {
					return err
				}
				return ethschnorr.Verify(dkg.share.Public(), dss.Message(), dss.sig)
			},
		},
	}
)

func runRound(t *testing.T, r round, members []*Manager) {
	buf := make([]PeerMessages, len(members))
	for i := range members {
		buf[i] = members[i].FlushMessages()
	}
	if r.tweak != nil {
		r.tweak(buf)
	}
	for i := range members {
		for j := range members {
			if i == j {
				continue
			}
			require.NoError(t, members[i].ProcessMessages(&buf[j]),
				"round=%s, i=%d, j=%d", r.name, i, j)
		}
	}
	if r.verify != nil {
		for i, m := range members {
			require.NoError(t, r.verify(m), "round=%s, i=%d", r.name, i)
		}
	}
}

func TestDkgBasic(t *testing.T) {
	numMembers := 2
	blockHeight := int64(1)

	members := make([]*Manager, numMembers)
	pubkeys := make([]Point, numMembers)
	for i := range members {
		members[i], pubkeys[i] = NewTestManager(t)
	}
	t.Logf("DKG with %d members", numMembers)

	for i, m := range members {
		require.NoError(t, m.StartDKG(blockHeight, pubkeys), "i=%d", i)
	}
	for _, r := range dkgRounds {
		runRound(t, r, members)
	}

	require.False(t, members[0].NeedDKG(pubkeys))
	require.Error(t, members[0].StartDKG(blockHeight, pubkeys))

	// add participants and refresh DKG
	for len(members) < 7 {
		blockHeight++

		m, pub := NewTestManager(t)
		members = append(members, m)
		pubkeys = append(pubkeys, pub)
		t.Logf("DKG with %d members", len(members))

		for i, m := range members {
			require.True(t, m.NeedDKG(pubkeys), "i=%d", i)
			require.NoError(t, m.StartDKG(blockHeight, pubkeys), "i=%d", i)
		}
		for _, r := range dkgRounds {
			runRound(t, r, members)
		}
	}
}

func TestSignatureBasic(t *testing.T) {
	numMembers := 7

	members := make([]*Manager, numMembers)
	pubkeys := make([]Point, numMembers)
	for i := range members {
		members[i], pubkeys[i] = NewTestManager(t)
	}

	for _, m := range members {
		require.NoError(t, m.StartDKG(1, pubkeys))
	}
	for _, r := range dkgRounds {
		runRound(t, r, members)
	}

	msg := big.NewInt(42)
	for _, m := range members {
		require.NoError(t, m.StartSignature(m.currentDkgID, msg))
	}
	runRound(t, dssRounds[0], members)

	require.Error(t, members[0].StartSignature(0, msg))

	require.NoError(t, members[0].StartDKG(2, pubkeys[:3]))
	require.Error(t, members[0].StartSignature(2, msg))
}
