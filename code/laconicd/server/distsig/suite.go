package distsig

import (
	"fmt"

	"git.vdb.to/cerc-io/chain-signatures/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
)

type (
	Scalar = kyber.Scalar
	Point  = kyber.Point

	PublicKey = secp256k1.PublicKey
)

var (
	suite     suites.Suite = secp256k1.NewBlakeKeccackSecp256k1()
	NewScalar              = suite.Scalar
	NewPoint               = suite.Point

	// Note: the compressed encoding of pubkeys used by Cosmos SDK and our library (based on chainlink)
	// are the same.  See gitlab.com/yawning/secp256k1-voi/secec
	SuitePublicKeyFromBytes = secp256k1.NewPublicKeyFromBytes
)

func KeyRecordToPoint(longterm *keyring.Record) (kyber.Point, error) {
	pubkey, err := longterm.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to access public key: %w", err)
	}
	suitePubkey, err := SuitePublicKeyFromBytes(pubkey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	return suitePubkey.Point()
}
