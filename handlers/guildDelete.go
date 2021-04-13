package handlers

import (
	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
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
