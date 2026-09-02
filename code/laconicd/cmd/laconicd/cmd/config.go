package cmd

import (
	"time"

	cmtcfg "github.com/cometbft/cometbft/config"
)

// initCometBFTConfig helps to override default CometBFT config values.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// display only warn logs by default for builtin modules except server, p2p, state
	cfg.LogLevel = "*:warn,server:info,p2p:info,state:info"
	cfg.LogLevel += ",auction:info,bond:info,registry:info,gql-server:info"

	// // TODO: understand full meaning of these settings
	cfg.Consensus.TimeoutPropose = 5000 * time.Millisecond
	// cfg.Consensus.TimeoutProposeDelta = 500 * time.Millisecond
	// cfg.Consensus.TimeoutPrevote = 1000 * time.Millisecond
	// cfg.Consensus.TimeoutPrevoteDelta = 500 * time.Millisecond
	// cfg.Consensus.TimeoutPrecommit = 1000 * time.Millisecond
	// cfg.Consensus.TimeoutPrecommitDelta = 500 * time.Millisecond

	// // start new block as soon as 2/3 precommits are received
	// cfg.Consensus.TimeoutCommit = 0 * time.Second

	cfg.Consensus.CreateEmptyBlocks = false

	// overwrite default pprof listen address
	cfg.RPC.PprofListenAddress = "localhost:6060"
	// use previous db backend
	cfg.DBBackend = "goleveldb"

	return cfg
}
