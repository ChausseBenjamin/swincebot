package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	"github.com/ChausseBenjamin/swincebot/internal/logging"
	discord "github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

// swinceSession stores state for multi-step swince interaction
type swinceSession struct {
	userID              string
	channelID           string
	guildID             string
	step                int
	taggedUsers         []int64
	selectedNominations map[int64]uuid.UUID // user -> nomination swince_id to fulfill
	videoURL            string
	nominations         map[int64]int64 // nominator -> nominated
	db                  *database.Queries
}

var swinceSessions = make(map[string]*swinceSession)

func swinceHandler(s *discord.Session, i *discord.InteractionCreate, db *database.Queries) {
	// Initialize session
	sessionID := i.Member.User.ID + "_" + i.ChannelID
	session := &swinceSession{
		userID:              i.Member.User.ID,
		channelID:           i.ChannelID,
		guildID:             i.GuildID,
		step:                1,
		selectedNominations: make(map[int64]uuid.UUID),
		nominations:         make(map[int64]int64),
		db:                  db,
	}
	swinceSessions[sessionID] = session

	// Step 1: Ask user to tag people present for the swince
	err := s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
		Type: discord.InteractionResponseChannelMessageWithSource,
		Data: &discord.InteractionResponseData{
			Content: "🎯 **Starting a new swince!**\n\nPlease tag all the people present for this swince (e.g., @user1 @user2 @user3):",
			Flags:   discord.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to send swince step 1 response", logging.ErrKey, err)
	}

	// Set up message listener for this session
	s.AddHandlerOnce(func(s *discord.Session, m *discord.MessageCreate) {
		if m.Author.ID == session.userID && m.ChannelID == session.channelID {
			handleSwinceMessage(s, m, sessionID)
		}
	})
}

func handleSwinceMessage(s *discord.Session, m *discord.MessageCreate, sessionID string) {
	session, exists := swinceSessions[sessionID]
	if !exists {
		return
	}

	ctx := context.Background()

	switch session.step {
	case 1: // Process tagged users
		// Extract user mentions from message
		userMentions := extractUserMentions(m.Content)
		if len(userMentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "❌ Please mention at least one user. Try again:")
			return
		}

		session.taggedUsers = userMentions
		session.step = 2

		// Check for unfulfilled nominations for each tagged user
		var nominationMessages []string

		for _, userID := range userMentions {
			nominations, err := session.db.GetUnfulfilledNominations(ctx, userID)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to get unfulfilled nominations", logging.ErrKey, err)
				continue
			}

			if len(nominations) == 0 {
				nominationMessages = append(nominationMessages, fmt.Sprintf("✅ <@%d> has no unfulfilled nominations", userID))
			} else if len(nominations) == 1 {
				// Auto-select the single nomination
				session.selectedNominations[userID] = nominations[0].SwinceID
				nominationMessages = append(nominationMessages, fmt.Sprintf("🎯 <@%d> has 1 unfulfilled nomination (auto-selected)", userID))
			} else {
				// Multiple nominations - need user to choose
				var optionsList []string
				for i, nom := range nominations {
					optionsList = append(optionsList, fmt.Sprintf("%d. Nomination from swince %s", i+1, nom.SwinceID.String()[:8]))
				}
				nominationMessages = append(nominationMessages,
					fmt.Sprintf("⚡ <@%d> has %d unfulfilled nominations:\n%s\nReply with the number to select:",
						userID, len(nominations), strings.Join(optionsList, "\n")))
			}
		}

		response := fmt.Sprintf("📋 **Tagged users processed:**\n%s\n\n📹 **Next step:** Please upload a video of the swince.",
			strings.Join(nominationMessages, "\n"))
		s.ChannelMessageSend(m.ChannelID, response)

		if len(session.selectedNominations) == len(session.taggedUsers) {
			// All nominations resolved, proceed to video upload
			session.step = 3
		}

	case 2: // Handle nomination selection (if needed)
		// Parse user selection for nominations
		// This is simplified - you'd want more robust parsing
		s.ChannelMessageSend(m.ChannelID, "✅ Nomination selection updated. Please upload a video of the swince.")
		session.step = 3

	case 3: // Process video upload
		if len(m.Attachments) == 0 {
			s.ChannelMessageSend(m.ChannelID, "❌ Please upload a video file.")
			return
		}

		// Check if attachment is a video
		attachment := m.Attachments[0]
		if !isVideoFile(attachment.Filename) {
			s.ChannelMessageSend(m.ChannelID, "❌ Please upload a video file (mp4, mov, avi, etc.).")
			return
		}

		// Store video URL for later processing
		session.videoURL = attachment.URL
		session.step = 4

		// Ask for nominations
		s.ChannelMessageSend(m.ChannelID, "🎯 **Video uploaded!** Now, for each person who participated, please tell me who they nominate next.\n\nFormat: @participant nominates @nominee")

	case 4: // Process nominations
		// Parse nominations from message
		nominations := parseNominations(m.Content)
		for nominator, nominee := range nominations {
			session.nominations[nominator] = nominee
		}

		// Check if all participants have made nominations
		if len(session.nominations) >= len(session.taggedUsers) {
			// Complete the swince process
			completeSwinceProcess(s, session, sessionID)
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Nominations recorded. Still need nominations from %d more participants.",
				len(session.taggedUsers)-len(session.nominations)))
		}
	}
}

func completeSwinceProcess(s *discord.Session, session *swinceSession, sessionID string) {
	ctx := context.Background()

	// Download video data
	videoData, err := downloadAttachment(session.videoURL)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to download video for storage", logging.ErrKey, err)
		s.ChannelMessageSend(session.channelID, "❌ Failed to process video. Please try the command again.")
		return
	}

	// Create new swince record
	swinceResult, err := session.db.CreateSwince(ctx, videoData)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create swince record", logging.ErrKey, err)
		s.ChannelMessageSend(session.channelID, "❌ Failed to save swince. Please try again.")
		return
	}

	// Create swinceur records and fulfill nominations
	for _, userID := range session.taggedUsers {
		nominatedUser := session.nominations[userID]

		// Create swinceur record
		err := session.db.CreateSwinceur(ctx, database.CreateSwinceurParams{
			SwinceID:            swinceResult.SwinceID,
			DiscordID:           userID,
			LateSwinceTax:       0.0, // You might want to calculate this based on timing
			Nominates:           nominatedUser,
			NominationFulfilled: 0.0, // This nomination is not yet fulfilled
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create swinceur record", logging.ErrKey, err)
			continue
		}

		// Fulfill existing nomination if selected
		if existingNomination, exists := session.selectedNominations[userID]; exists {
			err := session.db.FulfillNomination(ctx, database.FulfillNominationParams{
				SwinceID:  existingNomination,
				DiscordID: userID,
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to fulfill nomination", logging.ErrKey, err)
			}
		}
	}

	// Find swince channel and post the video
	swinceChannelID := findSwinceChannel(s, session.guildID)
	if swinceChannelID == "" {
		slog.WarnContext(ctx, "No swince channel found")
		swinceChannelID = session.channelID
	}

	// Create mentions for newly nominated people
	var newNominees []string
	for _, nomineeID := range session.nominations {
		newNominees = append(newNominees, fmt.Sprintf("<@%d>", nomineeID))
	}

	// Post final message in swince channel
	finalMessage := fmt.Sprintf("🎉 **New Swince Completed!**\n\n📹 Video: %s\n\n🎯 **Newly nominated:** %s\n\nSwince ID: `%s`\n\n**Video size:** %d bytes",
		session.videoURL,
		strings.Join(newNominees, " "),
		swinceResult.SwinceID.String()[:8],
		len(videoData))

	_, err = s.ChannelMessageSend(swinceChannelID, finalMessage)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to post swince completion message", logging.ErrKey, err)
	}

	// Confirm completion to user
	s.ChannelMessageSend(session.channelID, "✅ **Swince completed successfully!** Check the swince channel for the final post.")

	// Clean up session
	delete(swinceSessions, sessionID)
}

// Helper functions

func extractUserMentions(content string) []int64 {
	// Regex to match Discord user mentions <@!?123456789>
	re := regexp.MustCompile(`<@!?(\d+)>`)
	matches := re.FindAllStringSubmatch(content, -1)

	var userIDs []int64
	for _, match := range matches {
		if len(match) > 1 {
			if id := parseDiscordID(match[1]); id != 0 {
				userIDs = append(userIDs, id)
			}
		}
	}
	return userIDs
}

func parseNominations(content string) map[int64]int64 {
	// Parse format: @participant nominates @nominee
	nominations := make(map[int64]int64)

	// Simple regex to find "mentions nominates mentions" pattern
	re := regexp.MustCompile(`<@!?(\d+)>\s+nominates\s+<@!?(\d+)>`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			nominator := parseDiscordID(match[1])
			nominee := parseDiscordID(match[2])
			if nominator != 0 && nominee != 0 {
				nominations[nominator] = nominee
			}
		}
	}
	return nominations
}

func parseDiscordID(idStr string) int64 {
	// Convert string to int64 (Discord IDs are 64-bit integers)
	// You'd want proper error handling here
	var id int64
	fmt.Sscanf(idStr, "%d", &id)
	return id
}

func isVideoFile(filename string) bool {
	videoExtensions := []string{".mp4", ".mov", ".avi", ".mkv", ".webm", ".m4v"}
	filename = strings.ToLower(filename)
	for _, ext := range videoExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func downloadAttachment(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func findSwinceChannel(s *discord.Session, guildID string) string {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return ""
	}

	for _, channel := range channels {
		if channel.Type == discord.ChannelTypeGuildText &&
			(strings.Contains(strings.ToLower(channel.Name), "swince") ||
				strings.Contains(strings.ToLower(channel.Topic), "swince")) {
			return channel.ID
		}
	}
	return ""
}
