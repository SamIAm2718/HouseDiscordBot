package main

import (
	"flag"
	"fmt"
	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	token     string
	tokenPath string
	envName   string
	twitchClientId string
	twitchClientSecret string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&envName, "e", "", "Environment variable containing Bot Token")
	flag.StringVar(&tokenPath, "p", "", "Path to Bot Token")
	flag.StringVar(&twitchClientId, "etc", "", "Environment variable containing Twitch Client Id")
	flag.StringVar(&twitchClientSecret, "ets", "", "Environment variable containing Twitch Client Secret")
	flag.Parse()

	twitchClientId = os.Getenv("TWITCH_CLIENT_ID")
	twitchClientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
	// We process the most important flag to receive a token
	// The flags listed in order of importance are
	// t > e > p
	// If no flags are set the Bot exits with value ERR_NOFLAGS
	if len(token) > 0 {

	} else if len(envName) > 0 {
		token = os.Getenv(envName)
	} else if len(tokenPath) > 0 {
		rawToken, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			fmt.Println("Error reading token file,", err)
			os.Exit(constants.ERR_FILEREAD)
		}
		token = string(rawToken)
	} else {
		fmt.Println("Please specify a bot token using -t,")
		fmt.Println("an environment variable using -e,")
		fmt.Println("or a path to a bot token using -p.")
		os.Exit(constants.ERR_NOFLAGS)
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		os.Exit(constants.ERR_CREATEBOT)
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		os.Exit(constants.ERR_BOTOPEN)
	}

	go twitch.CheckIfHouseSlayerIsOnline(twitchClientId, twitchClientSecret)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\nBot is shutting down.")

	// Cleanly close down the Discord session.
	dg.Close()
}
