package ruleset

import (
	"fmt"
	"strings"

	"github.com/ChausseBenjamin/swincebot/internal/discord"
)

// LeaderboardEntry represents a user's position and score in the leaderboard
type LeaderboardEntry struct {
	User  discord.User
	Score int
	Rank  int
}

type Leaderboard []LeaderboardEntry

func (b Leaderboard) String() string {
	if len(b) == 0 {
		return "No entries in leaderboard"
	}

	var result strings.Builder
	result.WriteString("**Leaderboard**\n")

	for _, entry := range b {
		result.WriteString(fmt.Sprintf("%d. %s - %d pts\n",
			entry.Rank, entry.User.Nick, entry.Score))
	}

	return result.String()
}
