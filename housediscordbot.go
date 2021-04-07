package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/SamIAm2718/HouseDiscordBot/handlers"
	"github.com/SamIAm2718/HouseDiscordBot/twitch"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
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
			utils.Log.WithFields(logrus.Fields{"error": err}).Fatal("Token file could not be read")
		}
		token = string(rawToken)
	} else {
		utils.Log.Warning("No Flags specified.")
		utils.Log.Info("Loading bot token from the environment variable BOT_TOKEN.")
		token = os.Getenv("BOT_TOKEN")
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, errDiscord := discordgo.New("Bot " + token)
	if errDiscord != nil {
		utils.Log.WithFields(logrus.Fields{"error": errDiscord}).Fatal("Discord session could not be created.")
	}

	// Create a new Twitch session with client id, secret, and a path to saved data
	ts, errTwitch := twitch.New(os.Getenv("TWITCH_CLIENT_ID"), os.Getenv("TWITCH_CLIENT_SECRET"), "session1")
	if errTwitch != nil {
		utils.Log.WithFields(logrus.Fields{"error": errTwitch}).Error("Twitch session could not be created.")
	}

	utils.Log.Info("Bot is starting up.")

	dg.AddHandler(handlers.GuildCreate)
	dg.AddHandler(handlers.MessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	errDiscord = dg.Open()
	if errDiscord != nil {
		utils.Log.WithFields(logrus.Fields{"error": errDiscord}).Fatal("Could not establish connection to Discord.")
	}

	// Open a connection to twitch
	errTwitch = ts.Open()
	if errTwitch != nil {
		utils.Log.WithFields(logrus.Fields{"error": errTwitch}).Error("Could not establish connection to Twitch.")
	}

	// Start the twitch oracles
	go twitch.StartOracles(ts, dg)

	// Wait here until CTRL-C or other term signal is received.
	utils.Log.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly shut down the Twitch session
	utils.Log.Info("Twitch session is shutting down.")
	ts.Close()

	// Cleanly close down the Discord session.
	utils.Log.Info("Bot is shutting down.")
	dg.Close()
}
