package ruleset

import (
	"context"
	"fmt"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/discord"
)

// AllTimeScore calculates the cumulative score for a user across all seasons
func AllTimeScore(ctx context.Context, userID uint64) (int, error) {
	if len(rulesets) == 0 {
		return 0, fmt.Errorf("rulesets not initialized")
	}

	user := discord.User{ID: userID}
	totalScore := 0

	// Sum scores from all rulesets (seasons)
	for _, ruleset := range rulesets {
		score, err := ruleset.Score(ctx, user)
		if err != nil {
			// Continue accumulating even if one season fails
			continue
		}
		totalScore += score
	}

	return totalScore, nil
}

// AllTimeLeaderboard returns the all-time leaderboard across all seasons
func AllTimeLeaderboard(ctx context.Context, count int, db *database.ProtoDB) (Leaderboard, error) {
	if len(rulesets) == 0 {
		return nil, fmt.Errorf("rulesets not initialized")
	}

	// Get all unique users across all seasons
	userIDMap := make(map[uint64]bool)
	for _, ruleset := range rulesets {
		// Get leaderboard for this season to collect all users
		seasonLeaderboard, err := ruleset.Leaderboard(ctx, 0) // 0 means get all users
		if err != nil {
			continue // Skip seasons with errors
		}

		for _, entry := range seasonLeaderboard {
			userIDMap[entry.User.ID] = true
		}
	}

	// Calculate all-time scores for each user
	type userScore struct {
		user  discord.User
		score int
	}

	userScores := make([]userScore, 0, len(userIDMap))
	for userID := range userIDMap {
		user := discord.User{ID: userID}
		score, err := AllTimeScore(ctx, userID)
		if err != nil {
			continue // Skip users with calculation errors
		}

		userScores = append(userScores, userScore{user: user, score: score})
	}

	// Sort by score (descending)
	for i := 0; i < len(userScores)-1; i++ {
		for j := i + 1; j < len(userScores); j++ {
			if userScores[j].score > userScores[i].score {
				userScores[i], userScores[j] = userScores[j], userScores[i]
			}
		}
	}

	// Limit to requested count
	if count > 0 && count < len(userScores) {
		userScores = userScores[:count]
	}

	// Build leaderboard entries
	leaderboard := make(Leaderboard, len(userScores))
	for i, us := range userScores {
		leaderboard[i] = LeaderboardEntry{
			User:  us.user,
			Score: us.score,
			Rank:  i + 1,
		}
	}

	return leaderboard, nil
}
