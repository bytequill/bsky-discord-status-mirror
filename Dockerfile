FROM golang:1.22.7-alpine3.19

WORKDIR /app

COPY go.mod ./
COPY *.go ./
RUN go mod download


RUN go build -o bsky-discord

CMD [ "/bsky-discord" ]