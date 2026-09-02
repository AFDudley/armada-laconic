package gql

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cosmossdk.io/log"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/spf13/pflag"

	"git.vdb.to/cerc-io/laconicd/gql/schema"
	"git.vdb.to/cerc-io/laconicd/server"
)

const (
	ServerName = "gql"
)

type Server struct {
	logger     log.Logger
	router     chi.Router
	config     *Config
	httpServer *http.Server
}

// func New(logger log.Logger, cfg server.ConfigMap, clientCtx client.Context) (*Server, error) {
func New(clientCtx client.Context, cfg server.ConfigMap, logger log.Logger) (*Server, error) {
	s := &Server{
		logger: logger.With(log.ModuleKey, ServerName),
		router: chi.NewRouter(),
	}

	s.config = s.Config().(*Config)
	if len(cfg) > 0 {
		if err := server.UnmarshalSubConfig(cfg, s.Name(), &s.config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	s.router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		Debug:          false,
	}).Handler)

	srv := handler.NewDefaultServer(schema.NewExecutableSchema(schema.Config{Resolvers: &Resolver{
		ctx: clientCtx,
		// logFile: logFile,
	}}))

	s.router.Handle("/", PlaygroundHandler("/api"))

	if s.config.Playground {
		apiBase := s.config.PlaygroundAPIBase
		s.router.Handle("/webui", PlaygroundHandler(apiBase+"/api"))
		s.router.Handle("/console", PlaygroundHandler(apiBase+"/graphql"))
	}

	s.router.Handle("/api", srv)
	s.router.Handle("/graphql", srv)
	return s, nil
}

func (s *Server) Name() string {
	return ServerName
}

// Server configures and starts the GQL server.
func (s *Server) Start(ctx context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}

	if s.config.Playground {
		s.logger.Info(fmt.Sprintf("Connect to GraphQL playground URL at http://localhost:%d", s.config.Port))
	}
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.config.Port),
		Handler:           s.router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("Failed to start GraphQL server", "error", err)
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping GraphQL server")
	return s.httpServer.Shutdown(ctx)

}

func (s *Server) Config() any {
	if s.config == nil {
		return DefaultConfig()
	}
	return s.config
}

func (s *Server) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(s.Name(), pflag.ExitOnError)
	AddGQLFlags(flags)
	return flags
}
