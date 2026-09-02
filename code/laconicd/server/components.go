package server

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
)

// HasContextKey is implemented by servers that have a dedicated key under which they should be
// associated with a context.
type HasContextKey interface {
	ContextKey() string
}

type HasP2PReactors interface {
	P2PReactors() map[string]p2p.Reactor
}

// maps commands to server components they require
var cmdRequirements = map[*cobra.Command][]ServerComponent{}

// SetRequiredComponents sets required server components for a command
func SetRequiredComponents(cmd *cobra.Command, components ...ServerComponent) {
	for _, c := range components {
		cmdRequirements[cmd] = append(cmdRequirements[cmd], c)
		if startmod, ok := c.(HasStartFlags); ok {
			cmd.Flags().AddFlagSet(startmod.StartCmdFlags())
		}
	}
}

func RequiresComponent(cmd *cobra.Command, c string) bool {
	components, ok := cmdRequirements[cmd]
	if !ok {
		return false
	}
	for _, mod := range components {
		if mod.Name() == c {
			return true
		}
	}
	return false
}

type ComponentsCreator func(ConfigMap, client.Context, log.Logger, func(string) bool) ([]ServerComponent, error)

type ServerAux struct {
	components []ServerComponent
}

// AddComponents adds hooks to create, start and stop the components returned from the passed
// constructor, and ensures they are available on the command context.
// It returns a node.Option to additionally configure the CometBFT node.
func (s *ServerAux) AddComponents(cmd *cobra.Command, initComponents ComponentsCreator) {
	needsComponent := func(name string) bool {
		return RequiresComponent(cmd, name)
	}

	// manage component lifecycle
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		serverCtx := server.GetServerContextFromCmd(cmd)
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		s.components, err = initComponents(
			serverCtx.Viper.AllSettings(), clientCtx, serverCtx.Logger, needsComponent,
		)
		// add available context keys
		for _, mod := range s.components {
			if ckeymod, ok := mod.(HasContextKey); ok {
				cmdCtx := cmd.Context()
				if cmdCtx == nil {
					cmdCtx = context.Background()
				}
				cmd.SetContext(context.WithValue(cmdCtx, ckeymod.ContextKey(), mod))
			}
		}

		for _, mod := range s.components {
			if err := mod.Start(cmd.Context()); err != nil {
				return fmt.Errorf("failed to start server %s: %w", mod.Name(), err)
			}
		}
		return nil
	}

	startFn := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// if we do this in PostRunE, it won't be called if RunE errors
		defer func() {
			var err error
			for _, mod := range s.components {
				err = errors.Join(err, mod.Stop(cmd.Context()))
			}
			if err != nil {
				cmd.PrintErrln("failed to stop servers:", err)
			}
		}()
		return startFn(cmd, args)
	}
}

// AddReactors adds p2p Reactors for all configured components to a CometBFT node.
// When passed to StartCmdOptions this will be run in start(), after components are initialized in PreRun.
func (s *ServerAux) AddReactors(cmt *node.Node) {
	for _, mod := range s.components {
		if reactormod, ok := mod.(HasP2PReactors); ok {
			opt := node.CustomReactors(reactormod.P2PReactors())
			opt(cmt)
		}
	}
}
