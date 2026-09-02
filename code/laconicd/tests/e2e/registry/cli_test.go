package registry

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"git.vdb.to/cerc-io/laconicd/tests/e2e"
	"git.vdb.to/cerc-io/laconicd/testutil/network"
)

func TestRegistryE2ETestSuite(t *testing.T) {
	cfg := network.DefaultConfig(e2e.NewTestNetworkFixture)
	cfg.NumValidators = 1

	suite.Run(t, NewE2ETestSuite(cfg))
}
