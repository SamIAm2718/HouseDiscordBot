package twitch

import (
	"fmt"
	"os"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
)

type TwitchChannel string
type DiscordChannel string
type TwitchOracles map[TwitchChannel][]DiscordChannel

var (
	clientId     string
	clientSecret string
	Oracles      TwitchOracles
)

func init() {
	clientId = os.Getenv("TWITCH_CLIENT_ID")
	clientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
}

func MonitorChannel(t TwitchChannel, s *discordgo.Session) {
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

	var isOnline bool

	// TODO: refresh access token if expired
	for {
		if len(Oracles[t]) == 0 {
			fmt.Printf("No more channels monitoring for %v. Shutting down oracle.\n", string(t))
			delete(Oracles, t)
			return
		}

		resp, err := client.GetStreams(&helix.StreamsParams{
			UserLogins: []string{string(t)},
		})
		if err != nil {
			fmt.Println("Failed to query twitch", err)
		}

		if len(resp.Data.Streams) == 0 {
			// current offline
			if isOnline {
				isOnline = false
			}
		} else {
			// currently online
			if !isOnline {
				isOnline = true
				for _, d := range Oracles[t] {
					s.ChannelMessageSend(string(d), string(t)+" is online! Watch at http://twitch.tv/"+string(t))
				}
			}
		}
		time.Sleep(time.Second)
	}
}
