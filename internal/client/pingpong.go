package client

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	discord "github.com/bwmarrin/discordgo"
)

func ping(s *discord.Session, i *discord.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
		Type: discord.InteractionResponseChannelMessageWithSource,
		Data: &discord.InteractionResponseData{
			Content: "pong",
		},
	})
}

// /users command which lists all users in the current channel
func users(s *discord.Session, i *discord.InteractionCreate) {
	// Respond with a "thinking" message since fetching users might take time
	err := s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
		Type: discord.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discord.InteractionResponseData{},
	})
	slog.Error("failed to send initial response", logging.ErrKey, err)

	// Fetch members for the channel's guild
	members, err := s.GuildMembers(i.GuildID, "", 1000) // Limit to 1000 users
	if err != nil {
		msgErr := fmt.Sprintf("Error fetching users: %v", err)
		_, err = s.InteractionResponseEdit(i.Interaction, &discord.WebhookEdit{
			Content: &msgErr,
		})

	}

	// Build the response message
	var userList strings.Builder
	userList.WriteString("**Users in this channel:**\n")

	for _, member := range members {
		username := member.User.Username
		if member.Nick != "" {
			username = fmt.Sprintf("%s (%s)", member.Nick, username)
		}
		userList.WriteString(fmt.Sprintf("• %s\n", username))
	}

	userStr := userList.String()

	// Edit the response with the user list
	_, err = s.InteractionResponseEdit(i.Interaction, &discord.WebhookEdit{
		Content: &userStr,
	})
}

func options(s *discord.Session, i *discord.InteractionCreate) {
}
