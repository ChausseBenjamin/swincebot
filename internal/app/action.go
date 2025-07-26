package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	shutdownDone := make(chan struct{})

	var once sync.Once
	var db *database.ProtoDB
	gracefulShutdown := func() {}

	application := func() {
		var err error
		db, err = initApp(ctx, cmd)
		if err != nil {
			errAppChan <- err
			return
		}

		slog.InfoContext(ctx, "Starting application server")

		gracefulShutdown = func() {
			once.Do(func() {
				db.DB.Close()
				db.Queries.Close()
				slog.InfoContext(ctx, "Application shutdown")
				close(shutdownDone)
			})
		}

		slog.InfoContext(ctx, "Application listening")
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
			case <-time.After(cmd.Duration(FlagGraceTimeout)):
				slog.WarnContext(ctx, "Graceful shutdown delay exceeded, shutting down NOW!")
				if db != nil {
					go db.DB.Close()
					go db.Queries.Close()
				}
			case <-shutdownDone:
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

func initApp(ctx context.Context, cmd *cli.Command) (*database.ProtoDB, error) {
	globalConf := &util.ConfigStore{
		DBCacheSize: int(-cmd.Uint(FlagDBCacheSize)),
	}

	db, err := database.Setup(ctx, cmd.String(FlagDBPath), globalConf)
	if err != nil {
		return nil, err
	}

	return db, nil
}
