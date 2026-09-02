package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"git.vdb.to/cerc-io/laconicd/testutil/network"
)

// Reference: https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/testutil/cli/tx.go#L15

// CheckTxCode verifies that the transaction result returns a specific code
// Takes a network, wait for two blocks and fetch the transaction from its hash
func CheckTxCode(network *network.Network, clientCtx client.Context, txHash string, expectedCode uint32) error {
	// wait for 2 blocks
	for i := 0; i < 2; i++ {
		if err := network.WaitForNextBlock(); err != nil {
			return fmt.Errorf("failed to wait for next block: %w", err)
		}
	}

	cmd := authcli.QueryTxCmd()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{txHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
	if err != nil {
		return err
	}

	var response sdk.TxResponse
	if err := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response); err != nil {
		return err
	}

	if response.Code != expectedCode {
		return fmt.Errorf("expected code %d, got %d", expectedCode, response.Code)
	}

	return nil
}
