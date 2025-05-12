package app

import (
	"github.com/urfave/cli/v3"
)

const (
	AppName  = "swinceBot"
	AppUsage = "Your personal swince leaderboard manager ;)"
)

var version = "COMPILED"

func Command() *cli.Command {
	return &cli.Command{
		Name:    AppName,
		Usage:   AppUsage,
		Authors: []any{"Benjamin Chausse <benjamin@chausse.xyz>"},
		Version: version,
		Flags:   flags(),
		Action:  action,
	}
}
