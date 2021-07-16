package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/samuel-mokhtar/DiscordTwitchBot/twitch"
	"github.com/samuel-mokhtar/DiscordTwitchBot/utils"
)

func GuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		utils.Log.Debugf("Guild %v is unavailable.\n", event.ID)
		twitch.SetGuildUnavailable(event.ID)
		return
	}

	utils.Log.Debugf("Connected to guild %v.\n", event.ID)
	twitch.SetGuildActive(event.ID)
}
