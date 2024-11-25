FROM golang:1.22.7-alpine3.19

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN go build -o bsky-discord

CMD [ "/bsky-discord" ]