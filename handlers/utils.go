package handlers

import (
	"strings"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func isUserMod(ds *discordgo.Session, guildID string, user *discordgo.Member) bool {
	modID := getModRoleID(ds, guildID)

	if modID == "" {
		return false
	}

	for _, role := range user.Roles {
		if role == modID {
			return true
		}
	}

	return false
}

func getModRoleID(ds *discordgo.Session, guildID string) string {
	guildRoles, err := ds.GuildRoles(guildID)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{"error": err}).Error("Failed to get roles from guild.")
		return ""
	}

	for _, role := range guildRoles {
		if strings.EqualFold(role.Name, constants.ModRole) {
			return role.ID
		}
	}

	return ""
}
