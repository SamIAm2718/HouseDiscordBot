package main

import (
	"encoding/json"
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
	oracles   []twitch.TwitchOracle
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

	// Read twitch Oracle Data from JSONs
	readOraclesFromDisk(&oracles)
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		os.Exit(constants.ERR_CREATEBOT)
	}

	fmt.Println("Bot is starting up.")

	dg.AddHandler(handlers.GuildCreate)
	dg.AddHandler(handlers.MessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		os.Exit(constants.ERR_BOTOPEN)
	}

	// Register the twitch oracles

	for _, oracle := range oracles {
		go twitch.MonitorChannel(oracle, dg)
		fmt.Printf("Registering twitch oracle for, %+v\n", oracle)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	fmt.Println("\nBot is shutting down.")
	dg.Close()

	// Save twitch oracles for next session
	fmt.Println("Writing twitch oracles to disk.")
	writeOraclesToDisk(oracles)
}

func readOraclesFromDisk(o *[]twitch.TwitchOracle) {
	rawData, err := os.ReadFile("oracles.json")
	if err != nil {
		fmt.Println("Error reading oracles from json,", err)
		return
	}

	err = json.Unmarshal(rawData, o)
	if err != nil {
		fmt.Println("Error converting json to oracles,", err)
		return
	}
}

func writeOraclesToDisk(o []twitch.TwitchOracle) {
	b, err := json.Marshal(o)
	if err != nil {
		fmt.Println("Error converting oracles to json,", err)
		return
	}

	err = os.WriteFile("oracles.json", b, 0666)
	if err != nil {
		fmt.Println("Error writing oracles to disk,", err)
		return
	}
}
