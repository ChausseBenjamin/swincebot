package client

import (
	discord "github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	admins []int64
	*discord.Session
}

func New(token string, admins []int64) (*DiscordBot, error) {
	bot, err := discord.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	cmds := []*discord.ApplicationCommand{
		{
			Name:        "Basic Command",
			Description: "Just a description",
		},
	}
	bot.AddHandler(func(s *discord.Session, i *discord.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	return &DiscordBot{
		admins:  admins,
		Session: bot,
	}, nil

}
