package relay

import (
	"fmt"

	"git.vdb.to/cerc-io/laconicd/server"
)

type Config struct {
	// Target string `mapstructure:"target" toml:"target" comment:"The target peer to relay messages to"`
}

func DefaultConfig() *Config {
	return &Config{}
}

func UnmarshalConfig(cfg map[string]any) (*Config, error) {
	config := DefaultConfig()
	if err := server.UnmarshalSubConfig(cfg, serverName, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %T: %w", config, err)
	}
	return config, nil
}
