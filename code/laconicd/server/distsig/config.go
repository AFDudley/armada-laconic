package distsig

import (
	"fmt"

	"git.vdb.to/cerc-io/laconicd/server"
)

type Config struct {
	Enable      bool   `mapstructure:"enable" toml:"enable" comment:"Enable distributed Schnorr signatures"`
	LongtermKey string `mapstructure:"longterm-key" toml:"longterm-key" comment:"The long-term key used in distributed signatures"`
	// TODO: set this in genesis
	ThresholdRatio float64 `mapstructure:"threshold-ratio" toml:"threshold-ratio" comment:"The ratio of signers required to sign a message"`
}

func DefaultConfig() *Config {
	return &Config{
		Enable:         false,
		ThresholdRatio: 4. / 7,
	}
}

func UnmarshalConfig(cfg map[string]any) (*Config, error) {
	config := DefaultConfig()
	if err := server.UnmarshalSubConfig(cfg, componentName, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %T: %w", config, err)
	}
	return config, nil
}
