package testutil

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	modtestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func NewKeyring() keyring.Keyring {
	encCfg := modtestutil.MakeTestEncodingConfig()
	return keyring.NewInMemory(encCfg.Codec)
}
