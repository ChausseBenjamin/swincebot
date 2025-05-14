package client

import (
	"fmt"
	"log/slog"

	discord "github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	admins      []int64
	cmds        []*discord.ApplicationCommand
	cmdHandlers map[string]func(s *discord.Session, i *discord.InteractionCreate)
	*discord.Session
}

func New(token string, admins []int64) (*DiscordBot, error) {
	bot, err := discord.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	cmds := []*discord.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with pong",
		},
	}

	bot.Identify.Intents = discord.IntentsAllWithoutPrivileged
	// Command handlers map
	cmdHandlers := map[string]func(s *discord.Session, i *discord.InteractionCreate){
		"ping": func(s *discord.Session, i *discord.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discord.InteractionResponse{
				Type: discord.InteractionResponseChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: "pong",
				},
			})
		},
	}

	// Add interaction handler
	bot.AddHandler(func(s *discord.Session, i *discord.InteractionCreate) {
		if h, ok := cmdHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
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
	}, nil

}

// RegisterCommands must be called after the bot is initialized with `.Open()`
func (b *DiscordBot) RegisterCommands() ([]*discord.ApplicationCommand, error) {
	registeredCmds := make([]*discord.ApplicationCommand, len(b.cmds))

	for i, cmd := range b.cmds {
		registered, err := b.ApplicationCommandCreate(b.State.User.ID, "", cmd)
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
