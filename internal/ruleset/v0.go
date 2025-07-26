package ruleset

import (
	"context"
	"fmt"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/discord"
)

// v0 implements the initial ruleset for swincebot
// Point system:
// - Performing a swince: 1pt
// - Having a nomination fulfilled: 2pt (to nominator)
// - Fulfilling a nomination: 2pt (to nominee)
type v0 struct {
	db          *database.ProtoDB
	seasonIndex int
}

// setSeason sets the season index for this v0 ruleset instance
func (rs *v0) setSeason(seasonIndex int) {
	rs.seasonIndex = seasonIndex
}

func (rs v0) String() string { // TODO: make the points section a discordgo compatible table
	return `Ruleset:

	Ranking by the total amount of swinces alone wouldn't do enough justice to the effort put in by the people!
  Thus, multiple types of scores are kept and weights have been created.
  Each of these are combined into a total amount of points, which is what dictates your rank!

- Swince: **1pt**: Perform a Swince (and have it be submitted to the channel),
- Respond: **2pt**: You fulfilled a nomination
- Referall Bonus (Nomination response): **2pt**: Someone answered your nomination
`
}

func (rs *v0) Score(ctx context.Context, u discord.User) (int, error) {
	if rs.seasonIndex == seasonNotSet {
		return 0, fmt.Errorf("season not set for ruleset")
	}

	// Get season time range
	seasonStart, seasonEnd, err := rs.getSeasonTimeRange(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting season time range: %w", err)
	}

	userIDStr := fmt.Sprintf("%d", u.ID)

	// Get user swinces (1pt each)
	swinces, err := rs.db.GetUserSwinces(ctx, database.GetUserSwincesParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return 0, fmt.Errorf("getting user swinces: %w", err)
	}

	// Get user nominations that were fulfilled (2pts each)
	nominations, err := rs.db.GetUserNominations(ctx, database.GetUserNominationsParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return 0, fmt.Errorf("getting user nominations: %w", err)
	}

	// Get user fulfillments (2pts each)
	fulfillments, err := rs.db.GetUserFulfillments(ctx, database.GetUserFulfillmentsParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return 0, fmt.Errorf("getting user fulfillments: %w", err)
	}

	// Calculate total score: 1pt per swince + 2pts per nomination + 2pts per fulfillment
	totalScore := len(swinces) + (len(nominations) * 2) + (len(fulfillments) * 2)
	return totalScore, nil
}

func (rs *v0) ScoreStr(ctx context.Context, u discord.User) (string, error) {
	if rs.seasonIndex == seasonNotSet {
		return "", fmt.Errorf("season not set for ruleset")
	}

	// Get season time range
	seasonStart, seasonEnd, err := rs.getSeasonTimeRange(ctx)
	if err != nil {
		return "", fmt.Errorf("getting season time range: %w", err)
	}

	userIDStr := fmt.Sprintf("%d", u.ID)

	// Get individual components
	swinces, err := rs.db.GetUserSwinces(ctx, database.GetUserSwincesParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return "", fmt.Errorf("getting user swinces: %w", err)
	}

	nominations, err := rs.db.GetUserNominations(ctx, database.GetUserNominationsParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return "", fmt.Errorf("getting user nominations: %w", err)
	}

	fulfillments, err := rs.db.GetUserFulfillments(ctx, database.GetUserFulfillmentsParams{
		ParticipantID: userIDStr,
		Time:          seasonStart,
		Time_2:        seasonEnd,
	})
	if err != nil {
		return "", fmt.Errorf("getting user fulfillments: %w", err)
	}

	// Calculate individual scores
	swinceScore := len(swinces)
	nominationScore := len(nominations) * 2
	fulfillmentScore := len(fulfillments) * 2
	totalScore := swinceScore + nominationScore + fulfillmentScore

	return fmt.Sprintf("Total: %dpts, Swinces: %d (%dpts), Nominations: %d (%dpts), Responses: %d (%dpts)",
		totalScore, len(swinces), swinceScore, len(nominations), nominationScore, len(fulfillments), fulfillmentScore), nil
}

func (rs *v0) Leaderboard(ctx context.Context, count int) (Leaderboard, error) {
	if rs.seasonIndex == seasonNotSet {
		return nil, fmt.Errorf("season not set for ruleset")
	}

	// Get season time range
	seasonStart, seasonEnd, err := rs.getSeasonTimeRange(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting season time range: %w", err)
	}

	// Get all users who participated in this season
	userIDs, err := rs.db.GetAllUsersInSeason(ctx, database.GetAllUsersInSeasonParams{
		Time:   seasonStart,
		Time_2: seasonEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("getting users in season: %w", err)
	}

	// Calculate scores for each user
	type userScore struct {
		user  discord.User
		score int
	}

	userScores := make([]userScore, 0, len(userIDs))
	for _, userIDStr := range userIDs {
		// Parse user ID
		var userID uint64
		if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
			continue // Skip invalid user IDs
		}

		user := discord.User{ID: userID}
		score, err := rs.Score(ctx, user)
		if err != nil {
			continue // Skip users with score calculation errors
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

// NewV0 creates a new v0 ruleset with the provided database connection
func NewV0(db *database.ProtoDB) *v0 {
	return &v0{
		db:          db,
		seasonIndex: seasonNotSet,
	}
}

// getSeasonTimeRange returns the start and end time for this ruleset's season
func (rs *v0) getSeasonTimeRange(ctx context.Context) (time.Time, time.Time, error) {
	if rs.seasonIndex == seasonNotSet {
		return time.Time{}, time.Time{}, fmt.Errorf("season index not set")
	}

	// Get the start time for this season
	seasonStart, err := rs.db.GetSeasonStart(ctx, int64(rs.seasonIndex))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("getting season start: %w", err)
	}

	// Get the start time for the next season (which becomes our end time)
	nextSeasonStart, err := rs.db.GetNextSeasonStart(ctx, seasonStart)
	if err != nil {
		// If there's no next season, use current time as end
		return seasonStart, time.Now(), nil
	}

	return seasonStart, nextSeasonStart, nil
}
