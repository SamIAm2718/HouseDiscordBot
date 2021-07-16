# Discord Twitch Bot
## Running the Bot
To run the bot using go either run the command
```
go run discordtwitchbot.go
```
after setting the enviornment variable `BOT_TOKEN` to your Discord bot's token or run either of the commands
```
go run discordtwitchbot.go -t <Bot token>
go run discordtwitchbot.go -o <Path to file containing token>
```
if you don't want to set environment vairables (Note to use the Twitch functionality you will need to pass your Twitch app's client id through the environment variable TWITCH_CLIENT_ID and the Twitch app's secret through the enviornment variable TWITCH_CLIENT_SECRET). To run the project on Docker use the command

```
docker run -e BOT_TOKEN=<Bot Token> \
-e TWITCH_CLIENT_ID=<Twitch Client ID> \
-e TWITCH_CLIENT_SECRET=<Twitch Client Secret> \
--name <Container Name> samuel-mokhtar/discord-twitch-bot
```
To run the project as a kubernetes pod 
```
1. Create secret with 
    k create secret generic discordtwitchbot \
        --from-literal='bottoken=<bot token>' \
        --from-literal='twitchclientid=<twitch client id>' \
        --from-literal='twitchclientsecret=<twitch client secret>' \
    
2. kubectl apply -f k3sDiscordTwitchBot.yaml
```
Uses the repositories 
* https://github.com/bwmarrin/discordgo
* https://github.com/nicklaw5/helix
* https://github.com/sirupsen/logrus
* https://github.com/snowzach/rotatefilehook

## Using the Bot

To use the bot you can use the command
```
!twitch channel add <Twitch channel>
```
to register a Twitch channel to a Discord channel or
```
!twitch channel remove <Twitch channel>
```
To unregister a Twitch Channel from a Discord channel. You can use the command
```
!twitch channel list
```
to list the Twitch channels a Discord channel is monitoring.
