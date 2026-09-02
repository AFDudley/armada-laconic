package distsig

import (
	"fmt"

	"github.com/spf13/pflag"
)

func prefix(f string) string {
	return fmt.Sprintf("%s.%s", componentName, f)
}

var (
	FlagLongtermKey = prefix("longterm-key")
)

func AddNitroFlags(flags *pflag.FlagSet) {
	flags.String(FlagLongtermKey, "", "The long-term key used in distributed signatures")
}
