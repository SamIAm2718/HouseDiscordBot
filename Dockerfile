FROM golang:1.16

WORKDIR /go/src/housebot
COPY . .

RUN go install -v ./housediscordbot.go

RUN rm -r ./*

CMD ["housediscordbot"]