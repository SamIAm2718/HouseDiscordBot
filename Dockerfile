FROM golang:1.16

WORKDIR /go/src/app
COPY . .

RUN go install -v ./HouseDiscordBot.go

CMD ["HouseDiscordBot", "-e", "BOT_TOKEN"]