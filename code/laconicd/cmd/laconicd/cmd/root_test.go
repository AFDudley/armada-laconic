package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"git.vdb.to/cerc-io/laconicd/app"
	"git.vdb.to/cerc-io/laconicd/cmd/laconicd/cmd"
)

func TestInitCmd(t *testing.T) {
	rootCmd, err := cmd.NewRootCmd()
	require.NoError(t, err)
	rootCmd.SetArgs([]string{
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", app.DefaultNodeHome))
}

func TestHomeFlagRegistration(t *testing.T) {
	homeDir := "/tmp/foo"

	rootCmd, err := cmd.NewRootCmd()
	require.NoError(t, err)
	rootCmd.SetArgs([]string{
		"query",
		fmt.Sprintf("--%s", flags.FlagHome),
		homeDir,
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", app.DefaultNodeHome))

	result, err := rootCmd.Flags().GetString(flags.FlagHome)
	require.NoError(t, err)
	require.Equal(t, result, homeDir)
}

func TestHelpRequested(t *testing.T) {
	argz := [][]string{
		{"query", "--help"},
		{"query", "tx", "-h"},
		{"--help"},
		{"start", "-h"},
	}

	for _, args := range argz {
		rootCmd, err := cmd.NewRootCmd(args...)
		require.NoError(t, err)

		var out bytes.Buffer
		rootCmd.SetArgs(args)
		rootCmd.SetOut(&out)
		require.NoError(t, rootCmd.Execute())
		require.Contains(t, out.String(), args[0])
		require.Contains(t, out.String(), "--help")
		require.Contains(t, out.String(), "Usage:")
	}
}
