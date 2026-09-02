package nitro

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config with rate limiting",
			config: Config{
				Enable:             true,
				EthKey:             "test-eth-key",
				P2PRateLimitEnable: true,
				P2PRateLimitRate:   10.0,
				P2PRateLimitBurst:  20,
			},
			expectErr: false,
		},
		{
			name: "invalid rate - zero",
			config: Config{
				Enable:             true,
				EthKey:             "test-eth-key",
				P2PRateLimitEnable: true,
				P2PRateLimitRate:   0,
				P2PRateLimitBurst:  20,
			},
			expectErr: true,
		},
		{
			name: "invalid burst - zero",
			config: Config{
				Enable:             true,
				EthKey:             "test-eth-key",
				P2PRateLimitEnable: true,
				P2PRateLimitRate:   10.0,
				P2PRateLimitBurst:  0,
			},
			expectErr: true,
		},
		{
			name: "rate limiting disabled - no validation needed",
			config: Config{
				Enable:             true,
				EthKey:             "test-eth-key",
				P2PRateLimitEnable: false,
				P2PRateLimitRate:   0, // Invalid but should be ignored
				P2PRateLimitBurst:  0, // Invalid but should be ignored
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
