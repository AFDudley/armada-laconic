package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"time"

	"cosmossdk.io/log"
	sdk_slog "cosmossdk.io/log/slog"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/lmittmann/tint"

	"git.vdb.to/cerc-io/laconicd/server"
)

func createSlogLogger(cfg server.ConfigMap, out io.Writer) (log.Logger, error) {
	handler, err := createSlogHandler(cfg, out)
	if err != nil {
		return nil, err
	}
	return sdk_slog.NewCustomLogger(slog.New(handler)), nil
}

func createSlogHandler(cfg server.ConfigMap, out io.Writer) (slog.Handler, error) {
	var (
		format  string
		noColor bool
		level   string
	)
	if v, ok := cfg[flags.FlagLogFormat]; ok {
		format = v.(string)
	}
	if v, ok := cfg[flags.FlagLogNoColor]; ok {
		noColor = v.(bool)
	}
	if v, ok := cfg[flags.FlagLogLevel]; ok {
		level = v.(string)
	}

	logLvl, err := parseSlogLevel(level)
	if err != nil {
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevelWithParser(level, parseSlogLevel)
		if err != nil {
			return nil, err
		}
		out = log.NewFilterWriter(out, filterFunc)
	}

	var handler slog.Handler
	// if json format is requested, ignore no_color
	if format == flags.OutputFormatJSON {
		handler = slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level: logLvl,
		})
	} else {
		handler = tint.NewHandler(out, &tint.Options{
			Level:      logLvl,
			TimeFormat: time.RFC3339Nano,
			NoColor:    noColor,
		})
	}
	return handler, nil
}

func parseSlogLevel(l string) (slog.Level, error) {
	// Legal log level values are hardcoded into server/v2 command factory:
	// (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')
	// however, slog only has debug, info, warn, error
	switch l {
	case "debug", "trace":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "fatal", "panic":
		return slog.LevelError, nil
	case "disabled":
		return slog.Level(math.MaxInt), nil
	}
	return 0, fmt.Errorf("invalid log level: %s", l)
}
