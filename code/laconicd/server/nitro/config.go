package nitro

import (
	"errors"
	"fmt"

	"git.vdb.to/cerc-io/laconicd/server"
)

type Config struct {
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable Nitro state channel functionality"`

	EthKey string `mapstructure:"eth-key" toml:"eth-key" comment:"The private key used when interacting with the Ethereum chain and for our identity as a participant in the Nitro protocol."`
	// UseDistsig bool   `mapstructure:"use-distsig" toml:"use-distsig" comment:"Whether to use distributed signatures to authenticate Nitro actions."`

	// EthChainID    string
	EthURL        string `mapstructure:"eth-url" toml:"eth-url" comment:"The URL of the Ethereum node to connect to."`
	EthAuthToken  string `mapstructure:"eth-auth-token" toml:"eth-auth-token" comment:"The bearer token used for auth in requests to the Ethereum chain's RPC endpoint."`
	EthStartBlock uint64 `mapstructure:"eth-start-block" toml:"eth-start-block" comment:"Ethereum block number to start listening for Nitro Adjudicator events."`

	// TODO: move to module params?
	EthNaAddress  string `mapstructure:"eth-na-address" toml:"eth-na-address" comment:"Ethereum address of the Nitro Adjudicator contract."`
	EthVpaAddress string `mapstructure:"eth-vpa-address" toml:"eth-vpa-address" comment:"Ethereum address of the Virtual Payment App contract."`
	EthCaAddress  string `mapstructure:"eth-ca-address" toml:"eth-ca-address" comment:"Ethereum address of the Consensus App contract."`

	// P2P Rate Limiting
	P2PRateLimitEnable bool    `mapstructure:"p2p-rate-limit-enable" toml:"p2p-rate-limit-enable" comment:"Enable send-side rate limiting for P2P messages."`
	P2PRateLimitRate   float64 `mapstructure:"p2p-rate-limit-rate" toml:"p2p-rate-limit-rate" comment:"Maximum number of P2P messages per second (e.g., 10.0 for 10 messages/second)."`
	P2PRateLimitBurst  int     `mapstructure:"p2p-rate-limit-burst" toml:"p2p-rate-limit-burst" comment:"Maximum burst size for P2P rate limiter."`
}

func DefaultConfig() *Config {
	return &Config{
		Enable: true,
		// UseDistsig: false,
		// EthURL: "ws://127.0.0.1:8545",

		// Default rate limiting: 10 messages/second with burst of 20
		P2PRateLimitEnable: true,
		P2PRateLimitRate:   10.0,
		P2PRateLimitBurst:  20,
	}
}

func (c Config) Validate() error {
	if !c.Enable {
		return nil
	}
	if c.EthKey == "" {
		return errors.New("nitro.eth-key must be set")
	}
	// TODO With distsig enabled, eth-key will be a mutually exclusive setting, and authentication
	// for signing groups will be done using longterm key
	//
	// if c.UseDistsig && c.EthKey != "" {
	//	return errors.New("nitro.eth-key should not be set when distsig is enabled")
	// }

	// Validate rate limiting config
	if c.P2PRateLimitEnable {
		if c.P2PRateLimitRate <= 0 {
			return errors.New("nitro.p2p-rate-limit-rate must be positive when rate limiting is enabled")
		}
		if c.P2PRateLimitBurst <= 0 {
			return errors.New("nitro.p2p-rate-limit-burst must be positive when rate limiting is enabled")
		}
	}

	return nil
}

func UnmarshalConfig(cfg map[string]any) (*Config, error) {
	config := DefaultConfig()
	if len(cfg) > 0 {
		if err := server.UnmarshalSubConfig(cfg, serverName, config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %T: %w", config, err)
		}
	}
	return config, nil
}
