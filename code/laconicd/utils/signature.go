package utils

import (
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/statechannels/go-nitro/crypto"
)

func DecodeEthereumAddress(message []byte, sig string) (common.Address, error) {
	if len(sig) > 2 && sig[:2] == "0x" {
		sig = sig[2:]
	}

	signature := crypto.SplitSignature(common.Hex2Bytes(sig))
	return crypto.RecoverEthMessageSignerAddress(message, signature)
}

// TODO: use compressed pubkey encoding?
func DecodeEthereumPubKey(message []byte, sig string) ([]byte, error) {
	if len(sig) > 2 && sig[:2] == "0x" {
		sig = sig[2:]
	}

	signature := crypto.SplitSignature(common.Hex2Bytes(sig))
	return crypto.RecoverEthMessageSignerPubKey(message, signature)
}

func EthAddressFromPubKey(pubkey []byte) (common.Address, error) {
	ecdsaPubKey, err := ethcrypto.UnmarshalPubkey(pubkey)
	if err != nil {
		return common.Address{}, err
	}
	return ethcrypto.PubkeyToAddress(*ecdsaPubKey), nil
}
