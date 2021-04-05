package handlers

import (
	"fmt"
	"strings"

	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "housebot") {
		commandParams := strings.Split(m.Content, " ")[1:]

		if len(commandParams) > 0 {
			switch commandParams[0] {
			case "channel":
				commandChannel(s, m, commandParams[1:])
				return
			}
		}

		fmt.Println("Invalid Command Entered:", m.Content)
	}
}

func commandChannel(s *discordgo.Session, m *discordgo.MessageCreate, c []string) {
	if len(c) == 2 {
		switch c[0] {
		case "add":
			t := twitch.GetSession(s)
			if t != nil {
				twitchChannel := strings.ToLower(c[1])

				if t.ContainsOracle(twitchChannel, m.ChannelID) {
					s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel is already registered to this Discord channel.")
				} else {

					fmt.Println("Registering twitch oracle for", twitchChannel, "in channel", m.ChannelID)
					t.RegisterOracle(twitchChannel, m.ChannelID)

					_, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel successfully registered to this Discord channel.")
					if err != nil {
						fmt.Println("Error sending message,", err)
					}
				}
			}
			return
		case "remove":
			t := twitch.GetSession(s)
			if t != nil {
				twitchChannel := strings.ToLower(c[1])

				if !t.ContainsOracle(twitchChannel, m.ChannelID) {
					_, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel is not registered to this Discord channel.")
					if err != nil {
						fmt.Println("Error sending message,", err)
					}
				} else {
					fmt.Println("Unregistering twitch oracle for", twitchChannel, "in channel", m.ChannelID)
					t.UnregisterOracle(twitchChannel, m.ChannelID)

					_, err := s.ChannelMessageSend(m.ChannelID, twitchChannel+"'s Twitch channel successfully unregistered from this Discord channel.")
					if err != nil {
						fmt.Println("Error sending message,", err)
					}
				}
			}
			return
		default:
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Proper usage is housebot channel [add/remove] <Twitch Channel>")
	if err != nil {
		fmt.Println("Error sending message,", err)
	}
}
