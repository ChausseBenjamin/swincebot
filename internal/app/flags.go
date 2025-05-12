package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/urfave/cli/v3"
)

// Avoids string mismatches when calling cmd.String(), cmd.Int(), etc...
const (
	FlagDBPath       = "database"
	FlagGraceTimeout = "grace-timeout"
	FlagListenPort   = "port"
	FlagLogFormat    = "log-format"
	FlagLogLevel     = "log-level"
	FlagLogOutput    = "log-output"
	FlagSecretsPath  = "secrets-path"
	FlagDBCacheSize  = "database-cache-size"
)

func flags() []cli.Flag {
	return []cli.Flag{
		// Logging {{{
		&cli.StringFlag{
			Name:    FlagLogFormat,
			Aliases: []string{"f"},
			Value:   "plain",
			Usage:   "plain, json, none",
			Sources: cli.EnvVars("LOG_FORMAT"),
			Action:  validateLogFormat,
		},
		&cli.StringFlag{
			Name:    FlagLogOutput,
			Aliases: []string{"o"},
			Value:   "stdout",
			Usage:   "stdout, stderr, file",
			Sources: cli.EnvVars("LOG_OUTPUT"),
			Action:  validateLogOutput,
		},
		&cli.StringFlag{
			Name:    FlagLogLevel,
			Aliases: []string{"l"},
			Value:   "info",
			Usage:   "debug, info, warn, error",
			Sources: cli.EnvVars("LOG_LEVEL"),
			Action:  validateLogLevel,
		}, // }}}
		// discord Bot {{{
		&cli.DurationFlag{
			Name:    FlagGraceTimeout,
			Aliases: []string{"t"},
			Value:   3 * time.Second,
			Sources: cli.EnvVars("GRACEFUL_TIMEOUT"),
		}, // }}}
		// Database {{{
		&cli.UintFlag{
			Name:    FlagDBCacheSize,
			Value:   16000,
			Usage:   "Database cache to keep in memory (MB)",
			Sources: cli.EnvVars("DATABASE_CACHE_SIZE"),
		},
		&cli.StringFlag{
			Name:    FlagDBPath,
			Aliases: []string{"d"},
			Value:   "store.db",
			Usage:   "database file",
			Sources: cli.EnvVars("DATABASE_PATH"),
		}, // }}}
		// Service {{{
		&cli.StringFlag{
			Name:    FlagSecretsPath,
			Value:   "/etc/secrets",
			Usage:   "Directory containing necessary secrets (ca_certs, private keys, etc...)",
			Sources: cli.EnvVars("SECRETS_PATH"),
		}, // }}}
	}
}

func validateLogOutput(ctx context.Context, cmd *cli.Command, s string) error {
	switch s {
	case "stdout", "stderr":
		return nil
	default:
		// assume file
		f, err := os.OpenFile(s, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.ErrorContext(
				ctx,
				fmt.Sprintf("Error creating/accessing provided log file %s", s),
			)
			return err
		}
		defer f.Close() //nolint:errcheck
		return nil
	}
}

func validateLogLevel(ctx context.Context, cmd *cli.Command, s string) error {
	for _, lvl := range []string{"deb", "inf", "warn", "err"} {
		if strings.Contains(strings.ToLower(s), lvl) {
			return nil
		}
	}
	slog.ErrorContext(
		ctx,
		fmt.Sprintf("Unknown log level provided: %s", s),
	)
	return logging.ErrInvalidLevel
}

func validateLogFormat(ctx context.Context, cmd *cli.Command, s string) error {
	s = strings.ToLower(s)
	if s == "json" || s == "plain" || s == "none" {
		return nil
	}
	return nil
}

// func validateListenPort(ctx context.Context, cmd *cli.Command, p int64) error {
// 	if p < 1024 || p > 65535 {
// 		slog.ErrorContext(
// 			ctx,
// 			fmt.Sprintf("Out-of-bound port provided: %d", p),
// 		)
// 		return util.ErrOutOfBoundsPort
// 	}
// 	return nil
// }
