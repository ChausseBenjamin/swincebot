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
	FlagDBPath              = "database"
	FlagGraceTimeout        = "grace-timeout"
	FlagLogFormat           = "log-format"
	FlagLogLevel            = "log-level"
	FlagLogOutput           = "log-output"
	FlagSecretsPath         = "secrets-path"
	FlagDBCacheSize         = "database-cache-size"
	FlagDiscordServer       = "discord-server-id"
	FlagDiscordChannel      = "discord-channel-id"
	FlagConversationTimeout = "discord-conversation-timeout"
)

func flags() []cli.Flag {
	return []cli.Flag{
		// Discord {{{
		&cli.UintFlag{
			Name:     FlagDiscordServer,
			Usage:    "Server the bot is involved int (1 bot per discord server)",
			Sources:  cli.EnvVars("DISCORD_GUILD_ID"),
			Required: true,
		},
		&cli.UintFlag{
			Name:     FlagDiscordChannel,
			Usage:    "Channel where official bot communications occur",
			Sources:  cli.EnvVars("DISCORD_CHANNEL_ID"),
			Required: true,
		},
		&cli.DurationFlag{
			Name:    FlagConversationTimeout,
			Usage:   "How long before an active DM conversation gets cancelled due to inactivity",
			Sources: cli.EnvVars("DISCORD_CONVERSATION_TIMEOUT"),
			Value:   15 * time.Minute,
		}, // }}}
		// Logging {{{
		&cli.StringFlag{
			Name:    FlagLogFormat,
			Usage:   "plain, json, none",
			Value:   "plain",
			Sources: cli.EnvVars("LOG_FORMAT"),
			Action:  validateLogFormat,
		},
		&cli.StringFlag{
			Name:    FlagLogOutput,
			Usage:   "stdout, stderr, file",
			Value:   "stdout",
			Sources: cli.EnvVars("LOG_OUTPUT"),
			Action:  validateLogOutput,
		},
		&cli.StringFlag{
			Name:    FlagLogLevel,
			Usage:   "debug, info, warn, error",
			Value:   "info",
			Sources: cli.EnvVars("LOG_LEVEL"),
			Action:  validateLogLevel,
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
			Value:   "store.db",
			Usage:   "database file",
			Sources: cli.EnvVars("DATABASE_PATH"),
		}, // }}}
		// Service {{{
		&cli.DurationFlag{
			Name:    FlagGraceTimeout,
			Usage:   "Maximum time given to terminate active connections before being force-killed",
			Value:   3 * time.Second,
			Sources: cli.EnvVars("GRACEFUL_TIMEOUT"),
		},
		&cli.StringFlag{
			Name:    FlagSecretsPath,
			Usage:   "Directory containing necessary secrets (tokens, etc...)",
			Value:   "/etc/secrets",
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
