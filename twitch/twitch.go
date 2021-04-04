package twitch

import (
	"fmt"
	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/nicklaw5/helix"
	"os"
	"time"
)


func CheckIfHouseSlayerIsOnline(clientId string, clientSecret string) {
	client, err := helix.NewClient(&helix.Options{
		ClientID: clientId,
		ClientSecret:  clientSecret,
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
			if currentState == true {
				currentState = false
				// send message to discord
			}
		} else {
			// currently online
			fmt.Println("Robert online")
			if currentState == false {
				currentState = false
				// send message to discord
			}
		}
		time.Sleep(time.Second)
	}
}