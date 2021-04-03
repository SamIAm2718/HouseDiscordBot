package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token     string
	TokenPath string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TokenPath, "p", "", "Path to Bot Token")
	flag.Parse()

	if Token == "" {

		if TokenPath == "" {
			fmt.Println("Please specify a bot token using -t or a path to a bot token using -p.")
			os.Exit(1)
		}

		rawToken, err := ioutil.ReadFile(TokenPath)
		if err != nil {
			fmt.Println("Error reading token file:", err)
			os.Exit(1)
		}

		Token = string(rawToken)
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\nBot is shutting down.")

	// Cleanly close down the Discord session.
	dg.Close()
}
