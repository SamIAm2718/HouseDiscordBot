# HouseDiscordBot
## Running the Bot
To run the bot using go either run the command
```
go run HouseDiscordBot.go
```
after setting the enviornment variable `BOT_TOKEN` to your Discord bot's token or run either of the commands
```
go run HouseDiscordBot.go -t <Bot token>
go run HouseDiscordBot.go -o <Path to file containing token>
```
if you don't want to set environment vairables (Note to use the Twitch functionality you will need to pass your Twitch app's client id through the environment variable TWITCH_CLIENT_ID and the Twitch app's secret through the enviornment variable TWITCH_CLIENT_SECRET). To run the project on Docker use the command

```
docker run -e BOT_TOKEN=<Bot Token> \
-e TWITCH_CLIENT_ID=<Twitch Client ID> \
-e TWITCH_CLIENT_SECRET=<Twitch Client Secret> \
--name <Container Name> samiam2718/house-discord-bot
```
To run the project as a kubernetes pod 
```
1. Create secret with 
    k create secret generic housebot \
        --from-literal='bottoken=<bot token>' \
        --from-literal='twitchclientid=<twitch client id>' \
        --from-literal='twitchclientsecret=<twitch client secret>' \
    
2. kubectl apply -f k3sHouseBot.yaml
```
Uses the repository https://github.com/bwmarrin/discordgo 

## Using the Bot

To use the bot you can use the command
```
housebot add channel <Twitch channel>
```
to register a Twitch channel to a Discord channel or
```
housebot remove channel <Twitch channel>
```
To unregister a Twitch Channel from a Discord channel 