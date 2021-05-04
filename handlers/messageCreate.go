package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "housebot") {

		utils.Log.WithFields(logrus.Fields{
			"user":       m.Author.Username,
			"command":    m.Content,
			"channel_id": m.ChannelID,
			"server_id":  m.GuildID}).Info("Command recieved.")

		commandParams := strings.Split(m.Content, " ")[1:]

		if len(commandParams) > 0 {
			switch commandParams[0] {
			case "channel":
				go deleteUserMessageWithDelay(s, m, time.Second)
				if isUserMod(s, m.GuildID, m.Member) {
					commandChannel(s, m, commandParams[1:])
					return
				} else {
					utils.Log.Info("User ", m.Author.Username, " tried to issue a command without proper permissions.")
					return
				}
			}
		}

		utils.Log.WithFields(logrus.Fields{
			"user":       m.Author.Username,
			"command":    m.Content,
			"channel_id": m.ChannelID,
			"server_id":  m.GuildID}).Info("Invalid command.")
	}
}

func commandChannel(s *discordgo.Session, m *discordgo.MessageCreate, c []string) {
	if len(c) == 1 {
		switch c[0] {
		case "list":
			t := twitch.GetSession(s)
			mChannels := t.GetMonitoredChannels(m.ChannelID)
			listFields := []*discordgo.MessageEmbedField{}

			for i, channel := range mChannels {
				listField := &discordgo.MessageEmbedField{
					Name:   "Channel " + fmt.Sprint(i+1),
					Value:  channel,
					Inline: false,
				}

				listFields = append(listFields, listField)
			}

			listEmbed := &discordgo.MessageEmbed{
				Title:  "This Discord channel is monitoring",
				Fields: listFields,
			}

			_, err := s.ChannelMessageSendEmbed(m.ChannelID, listEmbed)
			if err != nil {
				utils.Log.WithError(err).Error("Failed to send message to Discord.")
			}
			return
		default:
		}
	} else if len(c) == 2 {
		switch c[0] {
		case "add":
			t := twitch.GetSession(s)
			twitchChannel := strings.ToLower(c[1])

			if err := t.RegisterChannel(twitchChannel, m.GuildID, m.ChannelID); err != nil {
				utils.Log.WithFields(logrus.Fields{
					"user":           m.Author.Username,
					"twitch_channel": twitchChannel,
					"channel_id":     m.ChannelID,
					"server_id":      m.GuildID,
					"error":          err}).Info("Failed to register channel.")

				if errors.Is(err, constants.ErrTwitchUserDoesNotExist) {
					m, err := s.ChannelMessageSend(m.ChannelID, "The Twitch channel "+twitchChannel+" does not exist.")
					if err != nil {
						utils.Log.WithError(err).Error("Failed to send message to Discord.")
					} else {
						go deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
					}
				} else if errors.Is(err, constants.ErrTwitchUserRegistered) {
					m, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel is already added to this Discord channel.")
					if err != nil {
						utils.Log.WithError(err).Error("Failed to send message to Discord.")
					} else {
						go deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
					}
				} else {
					m, err := s.ChannelMessageSend(m.ChannelID, "Error registering channel. Connection to twitch may be down.")
					if err != nil {
						utils.Log.WithError(err).Error("Failed to send message to Discord.")
					} else {
						go deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
					}
				}
			} else {
				utils.Log.WithFields(logrus.Fields{
					"user":           m.Author.Username,
					"twitch_channel": twitchChannel,
					"channel_id":     m.ChannelID,
					"server_id":      m.GuildID}).Info("Succeeded in registering channel.")

				m, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel successfully added to this Discord channel.")
				if err != nil {
					utils.Log.WithError(err).Error("Failed to send message to Discord.")
				} else {
					go deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
				}
			}
			return
		case "remove":
			t := twitch.GetSession(s)
			twitchChannel := strings.ToLower(c[1])

			if t.UnregisterChannel(twitchChannel, m.GuildID, m.ChannelID) {
				utils.Log.WithFields(logrus.Fields{
					"user":           m.Author.Username,
					"twitch_channel": twitchChannel,
					"channel_id":     m.ChannelID,
					"server_id":      m.GuildID}).Info("Succeeded in unregistering channel.")

				m, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel successfully removed from this Discord channel.")
				if err != nil {
					utils.Log.WithError(err).Error("Failed to send message to Discord.")
				} else {
					deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
				}
			} else {
				utils.Log.WithFields(logrus.Fields{
					"user":           m.Author.Username,
					"twitch_channel": twitchChannel,
					"channel_id":     m.ChannelID,
					"server_id":      m.GuildID}).Info("Failed to unregister channel.")

				m, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel is not added to this Discord channel.")
				if err != nil {
					utils.Log.WithError(err).Error("Failed to send message to Discord.")
				} else {
					go deleteBotMessageWithDelay(s, m, constants.DiscordMessageDeleteDelay)
				}

			}
			return
		default:
		}
	}

	mes, err := s.ChannelMessageSend(m.ChannelID, "Proper usage is:\n"+"housebot channel list\n"+"housebot channel [add/remove] <Twitch Channel>")
	if err != nil {
		utils.Log.WithError(err).Error("Failed to send message to Discord.")
	} else {
		go deleteBotMessageWithDelay(s, mes, constants.DiscordMessageDeleteDelay)
	}
}
