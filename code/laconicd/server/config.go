package server

// The utilities in this file are copied from cosmos-sdk server/v2

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type ConfigMap map[string]any

// ReadConfig returns a viper instance of the config file
func ReadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName("config")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", configPath, err)
	}

	v.SetConfigName("app")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	v.WatchConfig()

	return v, nil
}

// UnmarshalSubConfig unmarshals the given (sub) config from the main config (given as a map) into the target.
// If subName is empty, the main config is unmarshaled into the target.
func UnmarshalSubConfig(cfg map[string]any, subName string, target any) error {
	var sub any
	if subName != "" {
		if val, ok := cfg[subName]; ok {
			sub = val
		}
	} else {
		sub = cfg
	}

	// Create a new decoder with custom decoding options
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		Result:           target,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	// Decode the sub-configuration
	if err := decoder.Decode(sub); err != nil {
		return fmt.Errorf("failed to decode sub-configuration: %w", err)
	}

	return nil
}
