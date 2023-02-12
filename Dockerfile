# FROM golang:1.19.5
FROM ubuntu AS build

RUN apt-get update && apt-get -y upgrade && apt-get -y install ca-certificates golang git

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build .

FROM alpine

RUN apk --no-cache add ca-certificates gcompat

WORKDIR /usr/local/bin

COPY --from=build /usr/src/app/timid ./

CMD ["./timid"]
