package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		Token = getTokenFromPath(TokenPath)
	}
}

func getTokenFromPath(tp string) string {
	var (
		rawToken []byte
		err      error
	)

	for {
		rawToken, err = ioutil.ReadFile(tp)
		if err != nil {
			fmt.Println("Error reading token file,", err)
			fmt.Println("Attempting to read again in 5 seconds.")
			time.Sleep(5 * time.Second)
			continue
		} else if !(len(rawToken) > 0) {
			fmt.Println("Token file is empty")
			fmt.Println("Attempting to read again in 5 seconds.")
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}

	return string(rawToken)
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
