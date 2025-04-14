# Plex RSS Sync

This Go application regularly checks Plex RSS feeds and, upon detecting changes, triggers a forced scan in Sonarr and Radarr using their APIs.

## Features

- Fetches RSS feeds every 2 seconds (configurable via REFRESH_RATE).
- Detects changes and sends an API request to trigger a scan.
- Stores RSS feed states in a BoltDB database.

## Environment Variables

- `REFRESH_RATE` : Interval between checks (in seconds).
- `PLEX_RSS_URLS` : Comma-separated list of RSS feed URLs.
- `SONARR_URL` : Sonarr API URL.
- `SONARR_API_KEY` : Sonarr API key.
- `RADARR_URL` : Radarr API URL.
- `RADARR_API_KEY` : Radarr API key.

## Running the Application

1. Clone the repository
2. Set the environment variables
3. Run the application using go run main.go or via Docker

### Docker

```bash
docker build -t plex-rss-watcher .
docker run -e PLEX_RSS_URLS="https://example.com/rss1,https://example.com/rss2" -e SONARR_URL="http://localhost:8989" -e SONARR_API_KEY="your-sonarr-api-key" -e RADARR_URL="http://localhost:7878" -e RADARR_API_KEY="your-radarr-api-key" plex-rss-watcher
```
