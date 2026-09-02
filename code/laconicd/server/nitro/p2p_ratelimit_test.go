package nitro

import (
	"testing"

	"cosmossdk.io/log"
	nitrop2p "github.com/statechannels/go-nitro/node/engine/messageservice"
	nitrotypes "github.com/statechannels/go-nitro/types"
	"github.com/stretchr/testify/assert"
)

func TestP2PRateLimit(t *testing.T) {
	// Test with rate limiting enabled
	config := &Config{
		P2PRateLimitEnable: true,
		P2PRateLimitRate:   2.0, // 2 messages per second
		P2PRateLimitBurst:  2,   // burst of 2
	}

	reactor := newReactor(config, log.NewNopLogger())

	// Check rate limit stats
	enabled, limit, burst, tokens := reactor.GetRateLimitStats()
	assert.True(t, enabled, "Rate limiting should be enabled")
	assert.Equal(t, 2.0, limit, "Rate limit should be 2.0")
	assert.Equal(t, 2, burst, "Burst should be 2")
	assert.Equal(t, float64(2), tokens, "Initial tokens should equal burst size")

	// Test that UpdateRateLimit works
	reactor.UpdateRateLimit(5.0, 10)
	enabled, limit, burst, _ = reactor.GetRateLimitStats()
	assert.True(t, enabled)
	assert.Equal(t, 5.0, limit)
	assert.Equal(t, 10, burst)
}

func TestP2PRateLimitDisabled(t *testing.T) {
	// Test with rate limiting disabled
	config := &Config{
		P2PRateLimitEnable: false,
	}

	reactor := newReactor(config, log.NewNopLogger())

	// Check rate limit stats
	enabled, limit, burst, tokens := reactor.GetRateLimitStats()
	assert.False(t, enabled, "Rate limiting should be disabled")
	assert.Equal(t, 0.0, limit)
	assert.Equal(t, 0, burst)
	assert.Equal(t, 0.0, tokens)
}

func TestP2PRateLimitSendBehavior(t *testing.T) {
	// Test actual send behavior with rate limiting
	config := &Config{
		P2PRateLimitEnable: true,
		P2PRateLimitRate:   1000.0, // High rate for testing
		P2PRateLimitBurst:  1,      // Small burst
	}

	reactor := newReactor(config, log.NewNopLogger())
	msgService := newMessageService("test", reactor)

	// Create a dummy participant and message
	participant, _ := nitrotypes.ParseParticipant("0x0000000000000000000000000000000000000001")
	message := nitrop2p.Message{}

	// This should return an error because reactor is not running
	err := msgService.Send(participant, message)

	// We expect an error because reactor is not running
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reactor not running")
}
