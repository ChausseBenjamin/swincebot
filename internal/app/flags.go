package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/ChausseBenjamin/swincebot/internal/pb"
	"github.com/urfave/cli/v3"
)

// Avoids string mismatches when calling cmd.String(), cmd.Int(), etc...
const (
	FlagConfigPath       = "config"
	FlagDBPath           = "database"
	FlagDisableHTTPS     = "disable-https"
	FlagDisablePubSignup = "disable-public-signups"
	FlagGraceTimeout     = "grace-timeout"
	FlagListenPort       = "port"
	FlagLogFormat        = "log-format"
	FlagLogLevel         = "log-level"
	FlagLogOutput        = "log-output"
	FlagMaxUsers         = "max-users"
	FlagSecretsPath      = "secrets-path"
	FlagMinPasswdLen     = "min-password-length"
	FlagMaxPasswdLen     = "max-password-length"
	FlagAccessTokenTTL   = "access-token-time-to-live"
	FlagRefreshTokenTTL  = "refresh-token-time-to-live"
	FlagDBCacheSize      = "database-cache-size"
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
		// gRPC {{{
		&cli.IntFlag{
			Name:    FlagListenPort,
			Aliases: []string{"p"},
			Value:   1157, // list in leetspeak :P
			Sources: cli.EnvVars("LISTEN_PORT"),
			Action:  validateListenPort,
		},
		&cli.BoolFlag{ // TODO: Implement https
			Name:    FlagDisableHTTPS,
			Value:   false,
			Usage:   `Disable secure https communication. WARNING: Be very careful using this. Only do this if your server is behind a reverse proxy that already handles https for it and you trust all network communications on that network.`,
			Sources: cli.EnvVars("DISABLE_HTTPS"),
		},
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
		},
		&cli.UintFlag{
			Name:    FlagMaxUsers,
			Value:   25,
			Usage:   "Maximum number of users that can get created without admin intervention",
			Sources: cli.EnvVars("MAX_USERS"),
		},
		&cli.BoolFlag{
			Name:    FlagDisablePubSignup,
			Usage:   "Deactivate public (non admin-based) signups",
			Sources: cli.EnvVars("DISABLE_PUBLIC_SIGNUP"),
		},
		&cli.UintFlag{ // Not validated, you're dumb if you set a value < MinPasswdLen
			Name:    FlagMaxPasswdLen,
			Usage:   "Maximum password length the server can accept",
			Value:   144, // An OG tweet seems reasonable
			Sources: cli.EnvVars("MAX_PASSWORD_LENGTH"),
		},
		&cli.UintFlag{
			Name:    FlagMinPasswdLen,
			Usage:   "Minimum password length the server can accept",
			Value:   8,
			Sources: cli.EnvVars("MIN_PASSWORD_LENGTH"),
		},
		&cli.DurationFlag{
			Name:    FlagAccessTokenTTL,
			Usage:   "Duration of an access json web token (JWT)",
			Value:   15 * time.Minute,
			Sources: cli.EnvVars("JWT_ACCESS_TTL"),
		},
		&cli.DurationFlag{
			Name:    FlagRefreshTokenTTL,
			Usage:   "Duration of a refresh json web token (JWT)",
			Value:   24 * time.Hour,
			Sources: cli.EnvVars("JWT_REFRESH_TTL"),
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

func validateListenPort(ctx context.Context, cmd *cli.Command, p int64) error {
	if p < 1024 || p > 65535 {
		slog.ErrorContext(
			ctx,
			fmt.Sprintf("Out-of-bound port provided: %d", p),
		)
		return pb.ErrOutOfBoundsPort
	}
	return nil
}
