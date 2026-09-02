package system

import (
	"flag"
	"os"
	"testing"
	"time"

	"cosmossdk.io/systemtests"
)

var (
	Sut *systemtests.SystemUnderTest

	verbose    bool
	nodesCount = flag.Int("nodes", 4, "number of nodes in the cluster")
	blockTime  = 3 * time.Second // expected time between blocks
	// txTimeout  = 10 * time.Second // timeout for transactions to be included in a block
	txTimeout = 10 * time.Minute // timeout for transactions to be included in a block

	workDir string
)

func TestMain(m *testing.M) {
	// sdk systemtests expects this to be set for v2 server/runtime
	os.Setenv("COSMOS_BUILD_OPTIONS", "v2")

	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workDir = dir
	if verbose {
		println("Work dir: ", workDir)
	}

	Sut = systemtests.NewSystemUnderTest("laconicd", verbose, *nodesCount, blockTime)
	// Sut.SetTestnetInitializer(initer)

	Sut.SetupChain() // setup chain and keyring

	// run tests
	exitCode := m.Run()

	// postprocess
	Sut.StopChain()
	if verbose || exitCode != 0 {
		Sut.PrintBuffer()
	}

	os.Exit(exitCode)
}
