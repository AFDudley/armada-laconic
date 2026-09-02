package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	integrationTest "git.vdb.to/cerc-io/laconicd/tests/integration"
	types "git.vdb.to/cerc-io/laconicd/x/bond"
)

type KeeperTestSuite struct {
	suite.Suite
	integrationTest.TestFixture

	queryClient types.QueryClient
}

func (kts *KeeperTestSuite) SetupTest() {
	err := kts.TestFixture.Setup()
	kts.Require().NoError(err)

	qr := kts.App.QueryHelper()
	kts.queryClient = types.NewQueryClient(qr)
}

func TestBondKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
