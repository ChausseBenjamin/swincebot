package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/ChausseBenjamin/swincebot/internal/secrets"
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
	slog.InfoContext(ctx, "Starting swincebot server")

	errAppChan := make(chan error)
	shutdownDone := make(chan struct{}) // Signals when graceful shutdown is done

	var once sync.Once
	gracefulShutdown := func() {}
	brutalShutdown := func() {}

	application := func() {
		db, err := initApp(ctx, cmd)
		if err != nil {
			errAppChan <- err
			return
		}

		//nolint:errcheck
		gracefulShutdown = func() {
			once.Do(func() { // Ensure brutal shutdown isn't triggered later
				db.DB.Close()
				db.Queries.Close()
				slog.InfoContext(ctx, "Application shutdown")
				close(shutdownDone) // Signal that graceful shutdown is complete
			})
		}

		//nolint:errcheck
		brutalShutdown = func() {
			slog.WarnContext(ctx,
				"Graceful shutdown delay exceeded, shutting down NOW!",
			)
			db.DB.Close()
			db.Queries.Close()
		}

		port := fmt.Sprintf(":%d", cmd.Int(FlagListenPort))
		_, err = net.Listen("tcp", port)
		if err != nil {
			errAppChan <- err
			return
		}
		slog.InfoContext(ctx, "Server listening", "port", cmd.Int(FlagListenPort))

		// TODO: Make the discord bot start listening
		// if err := server.Serve(listener); err != nil {
		// 	errAppChan <- err
		// }
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

// TODO: Initialize discord API client here
func initApp(ctx context.Context, cmd *cli.Command) (*database.ProtoDB, error) {
	globalConf := &util.ConfigStore{
		DBCacheSize: int(-cmd.Uint(FlagDBCacheSize)),
	}

	_, err := secrets.NewDirVault(cmd.String(FlagSecretsPath))
	if err != nil {
		return nil, err
	}

	db, err := database.Setup(ctx, cmd.String(FlagDBPath), globalConf)
	if err != nil {
		return nil, err
	}

	return db, nil
}
