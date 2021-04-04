package twitch

import (
	"fmt"
	"os"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
)

var (
	clientId     string
	clientSecret string
)

type TwitchOracle struct {
	TwitchChannel  string
	DiscordChannel string
}

func init() {
	clientId = os.Getenv("TWITCH_CLIENT_ID")
	clientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
}

func MonitorChannel(t TwitchOracle, s *discordgo.Session) {
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

	var (
		currentState bool
	)
	// TODO: refresh access token if expired
	for {
		resp, err := client.GetStreams(&helix.StreamsParams{
			UserLogins: []string{"HouseSlayer"},
		})
		if err != nil {
			fmt.Println("Failed to query twitch", err)
		}
		if len(resp.Data.Streams) == 0 {
			fmt.Println("Robert offline")
			// current offline
			if currentState {
				currentState = false
				// send message to discord
			}
		} else {
			// currently online
			fmt.Println("Robert online")
			if !currentState {
				currentState = false
				// send message to discord
			}
		}
		time.Sleep(time.Second)
	}
}
