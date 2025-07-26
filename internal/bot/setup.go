package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ChausseBenjamin/swincebot/internal/discord"
	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(*discordgo.Session, *discordgo.InteractionCreate)

type Bot struct {
	discord         *discord.Client
	serverID        uint64
	channelID       uint64
	commandHandlers map[string]CommandHandler
}

func NewBot(ctx context.Context, discordClient *discord.Client, serverID, channelID uint64) (*Bot, error) {
	bot := &Bot{
		discord:   discordClient,
		serverID:  serverID,
		channelID: channelID,
	}

	if err := bot.registerCommands(ctx); err != nil {
		return nil, fmt.Errorf("registering commands: %w", err)
	}

	bot.registerHandlers()

	return bot, nil
}

func (b *Bot) registerCommands(ctx context.Context) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "swince",
			Description: "Submit a swince challenge",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "participants",
					Description: "Users who participated in the challenge",
					Required:    false,
				},
			},
		},
	}

	session := b.discord.Session()
	for _, cmd := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, strconv.FormatUint(b.serverID, 10), cmd)
		if err != nil {
			return fmt.Errorf("creating application command %s: %w", cmd.Name, err)
		}
		slog.InfoContext(ctx, "Registered slash command", "command", cmd.Name)
	}

	return nil
}

func (b *Bot) registerHandlers() {
	// Define command handlers map
	b.commandHandlers = map[string]CommandHandler{
		"swince": b.handleSwinceCommand,
		// Future commands can be added here:
		// "leaderboard": b.handleLeaderboardCommand,
		// "scores": b.handleScoresCommand,
	}

	session := b.discord.Session()
	session.AddHandler(b.handleInteraction)
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name

	if handler, exists := b.commandHandlers[commandName]; exists {
		handler(s, i)
	} else {
		slog.Warn("Unknown command received", "command", commandName)
	}
}

func (b *Bot) Close() error {
	return b.discord.Close()
}
