package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ChausseBenjamin/swincebot/internal/app"
	"github.com/ChausseBenjamin/swincebot/internal/logging"
)

func main() {
	cmd := app.Command()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("Program quit unexpectedly", logging.ErrKey, err)
		os.Exit(1)
	}
}
