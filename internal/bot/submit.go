package bot

import (
	"log/slog"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/bwmarrin/discordgo"
)

func swinceParticipantsArg() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        "participants",
		Description: "Users who participated in the challenge",
		Required:    false,
	}
}

func (b *Bot) handleSwinceCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	slog.Info("Swince command received", "user_id", userID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: ":beer: **Swince Challenge started!** Check your DMs to continue the process.\n\n:bulb: **You can type 'cancel' in the DM anytime to cancel the challenge.**",

			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		slog.Error("Failed to respond to swince command", logging.ErrKey, err, "user_id", userID)
	}
}
