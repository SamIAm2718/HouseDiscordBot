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

	if strings.HasPrefix(m.Content, "HouseBot") {
		commandParams := strings.Split(m.Content, " ")[1:]

		if len(commandParams) > 0 {
			switch commandParams[0] {
			case "add":
				commandAdd(s, m, commandParams[1:])
				return
			case "remove":
				commandRemove(s, m, commandParams[1:])
				return
			default:
			}
		}

		fmt.Println("Invalid Command Entered:", m.Content)
	}
}

func commandAdd(s *discordgo.Session, m *discordgo.MessageCreate, c []string) {
	if len(c) == 2 {
		switch c[0] {
		case "channel":

			t := twitch.TwitchChannel(c[1])

			for _, d := range twitch.Oracles[t] {
				if d == twitch.DiscordChannel(m.ChannelID) {
					_, err := s.ChannelMessageSend(m.ChannelID, c[1]+"'s twitch has already been registered to this channel.")
					if err != nil {
						fmt.Println("Error sending message,", err)
					}
					return
				}
			}

			fmt.Println("Registering twitch oracle for", c[1], "in channel", m.ChannelID)

			if twitch.Oracles[t] != nil {
				twitch.Oracles[t] = append(twitch.Oracles[t], twitch.DiscordChannel(m.ChannelID))
			} else {
				twitch.Oracles[t] = []twitch.DiscordChannel{twitch.DiscordChannel(m.ChannelID)}
				go twitch.MonitorChannel(t, s)
			}

			return
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Proper usage is HouseBot add channel <Twitch Channel>")
	if err != nil {
		fmt.Println("Error sending message,", err)
	}
}

func commandRemove(s *discordgo.Session, m *discordgo.MessageCreate, c []string) {
	if len(c) == 2 {
		switch c[0] {
		case "channel":
			t := twitch.TwitchChannel(c[1])
			for i, d := range twitch.Oracles[t] {
				if d == twitch.DiscordChannel(m.ChannelID) {
					twitch.Oracles[t][i] = twitch.Oracles[t][len(twitch.Oracles[t])-1]
					twitch.Oracles[t][len(twitch.Oracles[t])-1] = ""
					twitch.Oracles[t] = twitch.Oracles[t][:len(twitch.Oracles[t])-1]
				}
			}
			_, err := s.ChannelMessageSend(m.ChannelID, c[1]+"'s twitch successfully removed from this channel.")
			if err != nil {
				fmt.Println("Error sending message,", err)
			}
			return
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Proper usage is HouseBot remove channel <Twitch Channel>")
	if err != nil {
		fmt.Println("Error sending message,", err)
	}
}
