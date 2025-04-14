FROM golang:1.24.2-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -o plex-rss-sync . \
    && chmod +x plex-rss-sync \
    && mkdir -p ./db \
    && touch ./db/hash.db

CMD ["./plex-rss-sync"]
