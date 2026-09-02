package gql

import "github.com/spf13/pflag"

var (
	FlagEnable            = prefix("enable")
	FlagPlayground        = prefix("playground")
	FlagPlaygroundAPIBase = prefix("playground-api-base")
	FlagPort              = prefix("port")
	//FlagLogFile = prefix("log-file")
)

// AddGQLFlags adds flags for the GraphQL server.
func AddGQLFlags(flags *pflag.FlagSet) {
	flags.Bool(FlagEnable, false, "Enable the GraphQL server")
	flags.Bool(FlagPlayground, false, "Enable the GraphQL playground")
	flags.String(FlagPlaygroundAPIBase, "", "GraphQL API base path to use in the playground")
	flags.String(FlagPort, "9473", "Port to use for the GraphQL server.")
	// cmd.PersistentFlags().String(FlagLogFile, "", "File to tail for GraphQL 'getLogs' API")
}

func prefix(f string) string {
	return ServerName + "." + f
}
