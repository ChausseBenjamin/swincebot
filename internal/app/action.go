package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/client"
	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/ChausseBenjamin/swincebot/internal/util"
	"github.com/urfave/cli/v3"
)

func action(ctx context.Context, cmd *cli.Command) error {
	err := logging.Setup(
		cmd.String(FlagLogLevel),
		cmd.String(FlagLogFormat),
		cmd.String(FlagLogOutput),
	)
	if err != nil {
		slog.WarnContext(ctx, "Error(s) occurred during logger initialization",
			logging.ErrKey, err,
		)
	}

	errAppChan := make(chan error)
	shutdownDone := make(chan struct{}) // Signals when graceful shutdown is done

	var once sync.Once
	gracefulShutdown := func() {}
	brutalShutdown := func() {}

	application := func() {
		db, bot, err := initApp(ctx, cmd)
		if err != nil {
			errAppChan <- err
			return
		}

		if err := bot.Open(); err != nil {
			slog.Error("An error occured starting the discord bot", logging.ErrKey, err)
		}
		_, err = bot.RegisterCommands()
		if err != nil {
			slog.Error("Failed to register slash commands for the discord bot", logging.ErrKey, err)
		}
		slog.InfoContext(ctx, "Starting swincebot server")

		//nolint:errcheck
		gracefulShutdown = func() {
			once.Do(func() { // Ensure brutal shutdown isn't triggered later
				db.DB.Close()
				db.Queries.Close()
				// bot.UnregisterCommands(cmds)
				bot.Close()
				slog.InfoContext(ctx, "Application shutdown")
				close(shutdownDone) // Signal that graceful shutdown is complete
			})
		}

		//nolint:errcheck
		brutalShutdown = func() {
			slog.WarnContext(ctx,
				"Graceful shutdown delay exceeded, shutting down NOW!",
			)
			go db.DB.Close()
			go db.Queries.Close()
			go bot.Close()
		}

		slog.InfoContext(ctx, "Discord bot listening")
	}
	go application()

	stopChan := waitForTermChan()
	running := true
	for running {
		select {
		case errApp := <-errAppChan:
			if errApp != nil {
				slog.ErrorContext(ctx, "Application error", logging.ErrKey, errApp)
			}
			return errApp
		case <-stopChan:
			slog.InfoContext(ctx, "Shutdown requested")
			go gracefulShutdown()

			select {
			case <-time.After(cmd.Duration(FlagGraceTimeout)): // Timeout exceeded
				brutalShutdown()
			case <-shutdownDone: // If graceful shutdown is timely, exit normally
			}
			running = false
		}
	}
	return nil
}

func waitForTermChan() chan os.Signal {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	return stopChan
}

func initApp(ctx context.Context, cmd *cli.Command) (*database.ProtoDB, *client.DiscordBot, error) {
	globalConf := &util.ConfigStore{
		DBCacheSize: int(-cmd.Uint(FlagDBCacheSize)),
	}

	db, err := database.Setup(ctx, cmd.String(FlagDBPath), globalConf)
	if err != nil {
		return nil, nil, err
	}

	bot, err := client.New(cmd.String(FlagDiscordToken), cmd.IntSlice(FlagBotAdmins), db.Queries)
	if err != nil {
		return nil, nil, err
	}

	return db, bot, nil
}
