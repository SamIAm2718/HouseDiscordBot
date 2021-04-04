package handlers

import (
	"github.com/bwmarrin/discordgo"
)

func GuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}
}
