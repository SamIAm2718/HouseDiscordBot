FROM golang:1.16

WORKDIR /go/src/app
COPY . .

RUN go install -v ./HouseDiscordBot.go

EXPOSE 6463-6472

CMD HouseDiscordBot -t $(cat Token.txt)