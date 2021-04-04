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
	envName   string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TokenPath, "p", "", "Path to Bot Token")
	flag.StringVar(&envName, "e", "", "Environment variable containing Bot Token")
	flag.Parse()

	if len(Token) > 0 {

	} else if len(envName) > 0 {
		Token = os.Getenv(envName)
	} else if len(TokenPath) > 0 {
		Token = getTokenFromPath(TokenPath)
	} else {
		fmt.Println("Please specify a bot token using -t,")
		fmt.Println("an environment variable using -e,")
		fmt.Println("or a path to a bot token using -p.")
		os.Exit(1)
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
			fmt.Println("Token file is empty.")
			fmt.Println("Attempting to read again in 5 seconds.")
			time.Sleep(5 * time.Second)
			continue
		} else if !testToken(string(rawToken)) {
			fmt.Println("Token in file may be invalid.")
			fmt.Println("Attempting to read again in 5 seconds.")
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}

	return string(rawToken)
}

func testToken(t string) bool {
	dg, err := discordgo.New("Bot " + t)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return false
	}

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return false
	}

	time.Sleep(time.Second)

	dg.Close()

	return true
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
