# FROM golang:1.19.5
FROM ubuntu

RUN apt-get update && apt-get -y upgrade && apt-get -y install ca-certificates golang git

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build .

COPY timid /usr/local/bin

ENTRYPOINT ["timid"]
