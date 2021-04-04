# HouseDiscordBot

To run the bot either use the commands

```
go run HouseDiscordBot.go -t <Bot token>
go run HouseDiscordBot.go -e <Env variable containing token>
go run HouseDiscordBot.go -p <Path to token>
```

or run the project on Docker using the command

```
docker run -e BOT_TOKEN=<Bot Token> --name <Container Name> SamIAm2718/house-discord-bot
```

Uses the repository https://github.com/bwmarrin/discordgo 


To run as a kubernetes pod 

```
1. Create secret with k create secret generic housebot --from-literal='bottoken=<bot token>'
2. kubectl apply -f k3sHouseBot.yaml
```