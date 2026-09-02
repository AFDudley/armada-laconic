package server

// The utilities in this file are copied from cosmos-sdk server/v2

import (
	"context"

	"github.com/spf13/pflag"
)

// ServerComponent is a server component that can be started and stopped.
type ServerComponent interface {
	// Name returns the name of the server component.
	Name() string

	// Start starts the server component.
	Start(context.Context) error
	// Stop stops the server component.
	// Once Stop has been called on a server component, it may not be reused.
	Stop(context.Context) error
}

// HasStartFlags is a server component that has start flags.
type HasStartFlags interface {
	// StartCmdFlags returns server start flags.
	// Those flags should be prefixed with the server name.
	// They are then merged with the server config in one viper instance.
	StartCmdFlags() *pflag.FlagSet
}
