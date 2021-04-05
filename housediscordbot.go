package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/handlers"
	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	token     string
	tokenPath string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&tokenPath, "p", "", "Path to Bot Token")

	flag.Parse()

	// We process the most important flag to receive a token
	// The flags listed in order of importance are
	// t > e > p
	// If no flags are set the Bot exits with value ERR_NOFLAGS
	if len(token) > 0 {

	} else if len(tokenPath) > 0 {
		rawToken, err := os.ReadFile(tokenPath)
		if err != nil {
			fmt.Println("Error reading token file,", err)
			os.Exit(constants.ERR_FILEREAD)
		}
		token = string(rawToken)
	} else {
		fmt.Println("No Flags specified. Loading bot token from")
		fmt.Println("the environment variable BOT_TOKEN.")
		token = os.Getenv("BOT_TOKEN")
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, errDiscord := discordgo.New("Bot " + token)
	if errDiscord != nil {
		fmt.Println("error creating Discord session,", errDiscord)
		os.Exit(constants.ERR_CREATEBOT)
	}

	// Create a new Twitch session with client id, secret, and a path to saved data
	ts, errTwitch := twitch.New(os.Getenv("TWITCH_CLIENT_ID"), os.Getenv("TWITCH_CLIENT_SECRET"), "data")
	if errTwitch != nil {
		fmt.Println("Error starting Twitch session,", errTwitch)
	}

	fmt.Println("Bot is starting up.")

	dg.AddHandler(handlers.GuildCreate)
	dg.AddHandler(handlers.MessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	errDiscord = dg.Open()
	if errDiscord != nil {
		fmt.Println("error opening Discord connection,", errDiscord)
		os.Exit(constants.ERR_BOTOPEN)
	}

	// Open a connection to twitch
	errTwitch = ts.Open()
	if errTwitch != nil {
		fmt.Println("Error establishing Twitch connection,", errTwitch)
	}

	// Start the twitch oracles
	go twitch.StartOracles(ts, dg)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly shut down the Twitch session
	fmt.Println("\nTwitch session is shutting down.")
	ts.Close()

	// Cleanly close down the Discord session.
	fmt.Println("Bot is shutting down.")
	dg.Close()
}
