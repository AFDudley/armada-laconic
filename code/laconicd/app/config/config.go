package config

import (
	"fmt"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/pelletier/go-toml/v2"

	"git.vdb.to/cerc-io/laconicd/app/params"
	"git.vdb.to/cerc-io/laconicd/server/nitro"
	"git.vdb.to/cerc-io/laconicd/server/relay"
)

type LaconicAppConfig struct {
	serverconfig.Config `mapstructure:",squash"`
	customConfigs       `mapstructure:",squash"`
}

// the custom parts of the full config, so we can encode them separately
type customConfigs struct {
	Nitro nitro.Config `mapstructure:"nitro" toml:"nitro"`
	Relay relay.Config `mapstructure:"relay" toml:"relay"`
}

var DefaultConfigTemplate string

func init() {
	var err error
	DefaultConfigTemplate, err = createConfigTemplate()
	if err != nil {
		panic(err)
	}
}

func DefaultConfig() LaconicAppConfig {
	srvCfg := serverconfig.DefaultConfig()
	// In laconicd, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0" + params.CoinUnit

	// Now we set the custom config default values.
	customAppConfig := LaconicAppConfig{
		*srvCfg,
		customConfigs{
			Nitro: *nitro.DefaultConfig(),
			Relay: *relay.DefaultConfig(),
		},
	}
	return customAppConfig
}

func createConfigTemplate() (string, error) {
	defaultCustomConfigs := customConfigs{
		Nitro: *nitro.DefaultConfig(),
		Relay: *relay.DefaultConfig(),
	}
	b, err := toml.Marshal(defaultCustomConfigs)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n%s", serverconfig.DefaultConfigTemplate, b), nil
}
