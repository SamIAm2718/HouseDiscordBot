FROM golang:1.16

WORKDIR /go/src/discordtwitchbot
COPY . .

RUN go install -v ./discordtwitchbot.go

RUN rm -rfv ./*

CMD ["discordtwitchbot"]