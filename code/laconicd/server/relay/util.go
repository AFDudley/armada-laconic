package relay

import (
	"runtime"

	"git.vdb.to/cerc-io/laconicd/utils"
	"github.com/spf13/cobra"
)

func GetServerFromCmd(cmd *cobra.Command) (*Server, error) {
	s, err := utils.GetFromContext[Server](cmd.Context(), ServerContextKey)
	if err != nil {
		return nil, err
	}
	for !s.Ready() {
		s.logger.Debug("waiting for relay server")
		runtime.Gosched()
	}
	return s, nil
}
