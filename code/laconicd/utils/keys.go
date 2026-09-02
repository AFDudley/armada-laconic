package utils

import (
	"cosmossdk.io/core/address"
	sdkerrors "cosmossdk.io/errors"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// cosmos-sdk crypto/keyring/record.go:138
func extractPrivKeyFromLocal(rl *keyring.Record_Local) (cmtcrypto.PrivKey, error) {
	if rl.PrivKey == nil {
		return nil, keyring.ErrPrivKeyNotAvailable
	}

	priv, ok := rl.PrivKey.GetCachedValue().(sdkcrypto.PrivKey)
	if !ok {
		return nil, sdkerrors.Wrap(keyring.ErrCastAny, "PrivKey")
	}
	return &PrivKey{priv}, nil
}

func ExtractPrivateKey(r *keyring.Record) (cmtcrypto.PrivKey, error) {
	local := r.GetLocal()
	if local == nil {
		return nil, keyring.ErrPrivKeyExtr
	}
	return extractPrivKeyFromLocal(local)
}

func ExtractPrivateKeyByUid(kr keyring.Keyring, uid string) (cmtcrypto.PrivKey, error) {
	r, err := kr.Key(uid)
	if err != nil {
		return nil, err
	}
	return ExtractPrivateKey(r)
}

// GetKeyRecord tries to get a key by either address or uid
func GetKeyRecord(kr keyring.Keyring, from string, ac address.Codec) (*keyring.Record, error) {
	var k *keyring.Record
	addr, err := ac.StringToBytes(from)
	if err == nil {
		k, err = kr.KeyByAddress(sdk.AccAddress(addr))
		if err != nil {
			return nil, err
		}
	} else {
		k, err = kr.Key(from)
		if err != nil {
			return nil, err
		}
	}
	return k, nil
}

// PrivKey wraps the SDK PrivKey type to satisfy cometbft/crypto.PrivKey
type PrivKey struct{ sdkcrypto.PrivKey }

func (k PrivKey) PubKey() cmtcrypto.PubKey { return PubKey{k.PrivKey.PubKey()} }

func (k PrivKey) Equals(other cmtcrypto.PrivKey) bool {
	if otherPk, ok := other.(*PrivKey); ok {
		return k.PrivKey.Equals(otherPk.PrivKey)
	}
	return false
}

// PubKey wraps the SDK PubKey type to satisfy cometbft/crypto.PubKey
type PubKey struct{ sdkcrypto.PubKey }

func (k PubKey) Equals(other cmtcrypto.PubKey) bool {
	if otherPk, ok := other.(*PubKey); ok {
		return k.PubKey.Equals(otherPk.PubKey)
	}
	return false
}
