package nitro

import (
	"fmt"

	"github.com/spf13/pflag"
)

// start flags are prefixed with the server name
func prefix(f string) string {
	return fmt.Sprintf("%s.%s", serverName, f)
}

var (
	FlagEthKey        = prefix("eth-key")
	FlagEthURL        = prefix("eth-url")
	FlagEthStartBlock = prefix("eth-start-block")
	FlagEthAuthToken  = prefix("eth-auth-token")

	// FlagEthNaAddress  = prefix("eth-na-address")
	// FlagEthVpaAddress = prefix("eth-vpa-address")
	// FlagEthCaAddress  = prefix("eth-ca-address")
)

func AddFlags(f *pflag.FlagSet) {
	f.String(FlagEthKey, "", "name of private key to use when interacting with the Ethereum chain")
	f.String(FlagEthURL, "ws://127.0.0.1:8545", "URL of the Ethereum node to connect to")
	f.String(FlagEthAuthToken, "", "bearer token used for auth in requests to the Ethereum chain's RPC endpoint")
	f.Uint64(FlagEthStartBlock, 0, "Ethereum block number to start listening for Nitro Adjudicator events")
}
