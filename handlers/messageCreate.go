package handlers

import (
	"fmt"
	"strings"

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

		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Proper usage is HouseBot remove channel <Twitch Channel>")
	if err != nil {
		fmt.Println("Error sending message,", err)
	}
}
