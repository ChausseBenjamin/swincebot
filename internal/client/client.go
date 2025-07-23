package client

import (
	"fmt"
	"log/slog"

	"github.com/ChausseBenjamin/swincebot/internal/database"
	discord "github.com/bwmarrin/discordgo"
)

var testMinIntValue = 3.0

type DiscordBot struct {
	admins      []int64
	cmds        []*discord.ApplicationCommand
	cmdHandlers map[string]func(s *discord.Session, i *discord.InteractionCreate)
	db          *database.Queries
	*discord.Session
}

func New(token string, admins []int64, db *database.Queries) (*DiscordBot, error) {
	bot, err := discord.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	cmds := []*discord.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with pong",
		},
		{
			Name:        "users",
			Description: "Lists users in the channel",
		},
		{
			Name:        "swince",
			Description: "Start a new swince session",
		},
		{
			Name:        "options",
			Description: "Test command options",
			Options: []*discord.ApplicationCommandOption{
				{
					Type:        discord.ApplicationCommandOptionString,
					Name:        "string-option",
					Description: "String option",
					Required:    true,
				},
				{
					Type:        discord.ApplicationCommandOptionInteger,
					Name:        "integer-option",
					Description: "Integer option",
					MinValue:    &testMinIntValue,
					MaxValue:    10,
					Required:    true,
				},
				{
					Type:        discord.ApplicationCommandOptionNumber,
					Name:        "number-option",
					Description: "Float option",
					MaxValue:    10.1,
					Required:    true,
				},
				{
					Type:        discord.ApplicationCommandOptionBoolean,
					Name:        "bool-option",
					Description: "Boolean option",
					Required:    true,
				},
			},
		},
	}

	bot.Identify.Intents = discord.IntentsAllWithoutPrivileged
	// Command handlers map
	cmdHandlers := map[string]func(s *discord.Session, i *discord.InteractionCreate){
		"ping":    ping,
		"users":   users,
		"options": options,
	}

	// Add interaction handler
	bot.AddHandler(func(s *discord.Session, i *discord.InteractionCreate) {
		switch i.ApplicationCommandData().Name {
		case "ping":
			ping(s, i)
		case "users":
			users(s, i)
		case "swince":
			// We'll call the swince handler with database access
			swinceHandler(s, i, db)
		case "options":
			options(s, i)
		default:
			if h, ok := cmdHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})

	bot.AddHandler(func(s *discord.Session, r *discord.Ready) {
		slog.Info(fmt.Sprintf("Logged in as: %v#%v",
			s.State.User.Username,
			s.State.User.Discriminator,
		))
	})

	slog.Info("Bot Created", "bot", bot)
	return &DiscordBot{
		admins:      admins,
		Session:     bot,
		cmds:        cmds,
		cmdHandlers: cmdHandlers,
		db:          db,
	}, nil

}

// RegisterCommands must be called after the bot is initialized with `.Open()`
func (b *DiscordBot) RegisterCommands() ([]*discord.ApplicationCommand, error) {
	registeredCmds := make([]*discord.ApplicationCommand, len(b.cmds))

	for i, cmd := range b.cmds {
		registered, err := b.ApplicationCommandCreate(b.State.User.ID, "797979354638057492", cmd)
		if err != nil {
			return nil, fmt.Errorf("cannot create '%v' command '%v'", cmd.Name, err)
		}
		registeredCmds[i] = registered
	}
	return registeredCmds, nil
}

func (b *DiscordBot) UnregisterCommands(cmds []*discord.ApplicationCommand) error {
	for _, cmd := range cmds {
		if err := b.ApplicationCommandDelete(b.State.User.ID, "", cmd.ID); err != nil {
			return fmt.Errorf("cannot delete '%v' command: %v", cmd.Name, err)
		}
	}
	return nil
}
