package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/ChausseBenjamin/swincebot/internal/secrets"
	"github.com/bwmarrin/discordgo"
)

type Client struct {
	session   *discordgo.Session
	serverID  string
	channelID string
}

func NewClient(ctx context.Context, serverID, channelID uint64, vault secrets.SecretVault) (*Client, error) {
	token, err := vault.Get("discord_bot_token")
	if err != nil {
		return nil, fmt.Errorf("reading discord bot token: %w", err)
	}

	// Clean token of any whitespace/newlines
	cleanToken := strings.TrimSpace(token.String())
	if cleanToken == "" {
		return nil, fmt.Errorf("discord bot token is empty")
	}

	session, err := discordgo.New("Bot " + cleanToken)
	if err != nil {
		return nil, fmt.Errorf("creating discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuilds

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("opening discord session: %w", err)
	}

	return &Client{
		session:   session,
		serverID:  fmt.Sprintf("%d", serverID),
		channelID: fmt.Sprintf("%d", channelID),
	}, nil
}

func (c *Client) Close() error {
	return c.session.Close()
}

func (c *Client) Session() *discordgo.Session {
	return c.session
}

func (c *Client) ServerID() string {
	return c.serverID
}

func (c *Client) ChannelID() string {
	return c.channelID
}
