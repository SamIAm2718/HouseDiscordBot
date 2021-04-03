FROM golang:1.16

SHELL ["/bin/bash", "-c"]

WORKDIR /go/src/app
COPY . .

ARG token

RUN echo -ne ${token} > .token

RUN go install -v ./HouseDiscordBot.go

CMD ["HouseDiscordBot", "-p", "/go/src/app/.token"]