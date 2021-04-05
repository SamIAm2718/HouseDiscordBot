package twitch

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
)

type (
	TwitchChannel  string
	DiscordChannel string
	TwitchOracles  map[TwitchChannel][]DiscordChannel
	TwitchStates   map[TwitchChannel]bool
)

var (
	clientId     string
	clientSecret string
	Oracles      TwitchOracles
	States       TwitchStates
)

func init() {
	clientId = os.Getenv("TWITCH_CLIENT_ID")
	clientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
	Oracles = TwitchOracles{}
	States = TwitchStates{}
}

func StartOracles(s *discordgo.Session) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURI:  "http://localhost",
	})

	if err != nil {
		fmt.Println("could not open connection to twitch client, please retry", err)
		os.Exit(constants.ERR_TWITCHFAILURE)
	}

	// set access token
	resp, err := client.RequestAppAccessToken([]string{""})
	if err != nil {
		fmt.Println("could not open authenticate twitch connection", err)
		os.Exit(constants.ERR_TWITCHFAILURE)
	}
	client.SetAppAccessToken(resp.Data.AccessToken)

	// TODO: refresh access token if expired
	for {
		twitchChannels := []string{}

		for tc := range Oracles {
			if len(Oracles[tc]) == 0 {
				fmt.Printf("No more channels monitoring for %v. Shutting down oracle.\n", string(tc))
				delete(Oracles, tc)
				delete(States, tc)
			} else {
				twitchChannels = append(twitchChannels, string(tc))
			}
		}

		resp, err := client.GetStreams(&helix.StreamsParams{
			UserLogins: twitchChannels,
		})
		if err != nil {
			fmt.Println("Failed to query twitch", err)
		}

		for tc := range Oracles {
			foundMatch := false
			for _, ts := range resp.Data.Streams {
				foundMatch = strings.EqualFold(string(tc), ts.UserLogin)
				fmt.Println(string(tc), ts.UserLogin, foundMatch, ts.Type)

				if foundMatch {
					if ts.Type == "live" && !States[tc] {
						States[tc] = true
						for _, d := range Oracles[tc] {
							s.ChannelMessageSend(string(d), string(tc)+" is online! Watch at http://twitch.tv/"+string(tc))
						}
					}
					break
				}
			}

			if !foundMatch && States[tc] {
				States[tc] = false
				for _, d := range Oracles[tc] {
					s.ChannelMessageSend(string(d), string(tc)+" is now offline!")
				}
			}
		}
		time.Sleep(2 * time.Minute)
	}
}
