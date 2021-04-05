# HouseDiscordBot

To run the bot using go either run the command
```
go run HouseDiscordBot.go
```
after setting the enviornment variable `BOT_TOKEN` to your Discord bot's token or run either of the commands
```
go run HouseDiscordBot.go -t <Bot token>
go run HouseDiscordBot.go -o <Path to file containing token>
```
if you don't want to set environment vairables. To run the project on Docker use the command

```
docker run -e BOT_TOKEN=<Bot Token> \
-e TWITCH_CLIENT_ID=<Twitch Client ID> \
-e TWITCH_CLIENT_SECRET=<Twitch Client Secret> \
--name <Container Name> SamIAm2718/house-discord-bot
```
To run the project as a kubernetes pod 
```
1. Create secret with k create secret generic housebot --from-literal='bottoken=<bot token>'
2. kubectl apply -f k3sHouseBot.yaml
```
Uses the repository https://github.com/bwmarrin/discordgo 
