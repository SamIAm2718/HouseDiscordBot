package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/samuel-mokhtar/DiscordTwitchBot/twitch"
	"github.com/samuel-mokhtar/DiscordTwitchBot/utils"
)

func GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable {
		utils.Log.Debugf("Guild %v is unavailable.\n", event.ID)
		twitch.SetGuildUnavailable(event.ID)
		return
	}

	utils.Log.Debugf("Removed from guild %v.\n", event.ID)
	twitch.SetGuildInactive(event.ID)
}
