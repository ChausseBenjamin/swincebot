package discord

import (
	"fmt"
	"log/slog"
	"strconv"
)

type User struct {
	ID uint64
	// Nickname in this specific server (not generic @)
	Nick string
}

func (c *Client) GetNick(userID uint64) (string, error) {
	member, err := c.session.GuildMember(c.serverID, fmt.Sprintf("%d", userID))
	if err != nil {
		return "", fmt.Errorf("getting guild member: %w", err)
	}

	if member.Nick != "" {
		return member.Nick, nil
	}
	return member.User.Username, nil
}

func (c *Client) GetMembers() ([]User, error) {
	members, err := c.session.GuildMembers(c.serverID, "", 1000)
	if err != nil {
		return nil, fmt.Errorf("getting guild members: %w", err)
	}

	users := make([]User, 0, len(members))
	for _, member := range members {
		if member.User.Bot {
			continue
		}

		userID, err := strconv.ParseUint(member.User.ID, 10, 64)
		if err != nil {
			slog.Warn("failed to parse user ID", "user_id", member.User.ID, "error", err)
			continue
		}

		nick := member.Nick
		if nick == "" {
			nick = member.User.Username
		}

		users = append(users, User{
			ID:   userID,
			Nick: nick,
		})
	}

	return users, nil
}
