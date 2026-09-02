package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

// GetCmdList queries all records.
func GetCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List records.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the records.
Example:
$ %s query %s list
`,
				version.AppName, registrytypes.ModuleName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := registrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.Records(cmd.Context(), &registrytypes.QueryRecordsRequest{})
			if err != nil {
				return err
			}

			recordsList := res.GetRecords()
			records := make([]registrytypes.ReadableRecord, len(recordsList))
			for i, record := range res.GetRecords() {
				records[i] = record.ToReadableRecord()
			}
			bytesResult, err := json.Marshal(records)
			if err != nil {
				return err
			}
			return clientCtx.PrintBytes(bytesResult)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
