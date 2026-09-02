package gql

func DefaultConfig() *Config {
	return &Config{
		Enable:            false,
		Playground:        false,
		PlaygroundAPIBase: "",
		Port:              9473,
		// LogFile:           "",
	}
}

type Config struct {
	Enable            bool   `mapstructure:"enable" toml:"enable" comment:"Enable the GraphQL server."`
	Playground        bool   `mapstructure:"playground" toml:"playground" comment:"Enable the GraphQL playground."`
	PlaygroundAPIBase string `mapstructure:"playground-api-base" toml:"playground-api-base" comment:"GraphQL API base path to use in the playground."`
	Port              uint   `mapstructure:"port" toml:"port" comment:"Port to use for the GraphQL server."`
	// LogFile           string `mapstructure:"log-file" toml:"log-file" comment:"File to tail for the 'getLogs' API."`
}
