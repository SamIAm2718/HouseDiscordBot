# HouseDiscordBot

To run the bot either use the commands

```
go run HouseDiscordBot.go -t <Bot Token>
go run HouseDiscordBot.go -p <Path to Token>
```

or build the project using Docker with the command

```
docker build -t house-discord-bot . --build-arg token=<Bot Token>
```

Uses the repository https://github.com/bwmarrin/discordgo 
