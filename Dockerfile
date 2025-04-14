FROM golang:1.24.2-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go build -ldflags "-s -w" -a -o plex-rss-watcher . \
    && chmod +x plex-rss-watcher

CMD ["./plex-rss-watcher"]
