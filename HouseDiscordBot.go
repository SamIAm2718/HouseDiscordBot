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

	// Read Data from JSONs
	readJSONFromDisk("data", "twitchstates.json", &twitch.States)
	readJSONFromDisk("data", "twitchoracles.json", &twitch.Oracles)
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

	for k := range twitch.Oracles {
		go twitch.MonitorChannel(k, dg)
		fmt.Println("Registering saved twitch oracles for", k)
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
	fmt.Println("Writing data to disk.")
	writeJSONToDisk("data", "twitchstates.json", twitch.States)
	writeJSONToDisk("data", "twitchoracles.json", twitch.Oracles)
}

func readJSONFromDisk(filePath string, fileName string, o interface{}) {
	if filePath != "" {
		fileName = "/" + fileName
	}

	rawData, err := os.ReadFile(filePath + fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(filePath+fileName, "does not exist. Will be created on close.")
			return
		} else {
			fmt.Println("Error reading from json,", err)
			return
		}
	}

	err = json.Unmarshal(rawData, o)
	if err != nil {
		fmt.Println("Error converting json,", err)
		return
	}
}

func writeJSONToDisk(filePath string, fileName string, o interface{}) {
	if filePath != "" {
		fileName = "/" + fileName
	}

	b, err := json.Marshal(o)
	if err != nil {
		fmt.Println("Error converting json,", err)
		return
	}

	err = os.WriteFile(filePath+fileName, b, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Directory", filePath, "does not exist. Creating directory")

			err = os.Mkdir(filePath, 0755)
			if err != nil {
				fmt.Println("Error creating directory,", err)
				return
			}

			err = os.WriteFile(filePath+fileName, b, 0666)
			if err != nil {
				fmt.Println("Error writing to disk,", err)
				return
			}

		} else {
			fmt.Println("Error writing to disk,", err)
			return
		}
	}
}
