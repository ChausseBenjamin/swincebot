package ruleset

import (
	"context"
	"fmt"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/discord"
)

const seasonNotSet = -1

// Ruleset defines the interface for scoring and leaderboard generation
// Each ruleset implementation represents a different season's scoring rules
type Ruleset interface {
	// String returns a human-readable explanation of the ruleset
	String() string

	// Calc calculates the score for a specific user under this ruleset
	Score(ctx context.Context, u discord.User) (int, error)

	// ScoreStr returns a "fancier" score string for a given ruleset
	// ex: "Total: xPts, Nominations: 4 (yPts), Responses 3 (zPtz)"
	ScoreStr(ctx context.Context, u discord.User) (string, error)

	// Leaderboard returns the top 'count' users with their scores and rankings
	Leaderboard(ctx context.Context, count int) (Leaderboard, error)

	// setSeason sets the season index for this ruleset instance (called at startup)
	setSeason(seasonIndex int)
}

// rulesets is the global slice - index matches season number
// index 0 being the 0th season before the first season timestamp
var rulesets []Ruleset

// InitializeRulesets initializes all rulesets with their season indices
func InitializeRulesets(db *database.ProtoDB) {
	rulesets = []Ruleset{
		&v0{db: db, seasonIndex: seasonNotSet},
		// Future seasons can be pre-allocated here
	}

	// Set season index for each ruleset
	for i, ruleset := range rulesets {
		ruleset.setSeason(i)
	}
}

// Get returns the appropriate ruleset for a given timestamp
// Rulesets should be initialized at startup, so no context is needed
func Get(t time.Time) (Ruleset, error) {
	if len(rulesets) == 0 {
		return nil, fmt.Errorf("rulesets not initialized")
	}

	// Default to the first (v0) ruleset for now
	// TODO: Implement proper season detection based on timestamp and database
	return rulesets[0], nil
}

// GetWithDB returns the appropriate ruleset for a given timestamp using database queries
func GetWithDB(ctx context.Context, t time.Time, db *database.ProtoDB) (Ruleset, error) {
	if len(rulesets) == 0 {
		return nil, fmt.Errorf("rulesets not initialized")
	}

	// Get the season ID for the given timestamp
	seasonID, err := db.GetSeasonID(ctx, t)
	if err != nil {
		// If no season found, default to v0
		return rulesets[0], nil
	}

	// Ensure we don't go out of bounds
	if int(seasonID) >= len(rulesets) {
		return rulesets[len(rulesets)-1], nil
	}

	return rulesets[seasonID], nil
}
