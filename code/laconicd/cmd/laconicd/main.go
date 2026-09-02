package main

import (
	"errors"
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"git.vdb.to/cerc-io/laconicd/app"
	_ "git.vdb.to/cerc-io/laconicd/app/params" // import for side-effects
	"git.vdb.to/cerc-io/laconicd/cmd/laconicd/cmd"
)

func main() {
	// reproduce default cobra behavior so that eager parsing of flags is possible.
	// see: https://github.com/spf13/cobra/blob/e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec/command.go#L1082
	args := os.Args[1:]
	rootCmd, err := cmd.NewRootCmd(args...)
	if err != nil {
		if _, pErr := fmt.Fprintln(os.Stderr, err); pErr != nil {
			panic(errors.Join(err, pErr))
		}
		os.Exit(1)
	}
	if err := svrcmd.Execute(rootCmd, cmd.EnvPrefix, app.DefaultNodeHome); err != nil {
		if _, pErr := fmt.Fprintln(rootCmd.OutOrStderr(), err); pErr != nil {
			panic(errors.Join(err, pErr))
		}
		os.Exit(1)
	}
}
