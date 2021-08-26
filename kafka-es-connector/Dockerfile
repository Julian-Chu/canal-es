FROM golang:1.16

WORKDIR /app

RUN apt-get update && apt-get -yq install kafkacat
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
