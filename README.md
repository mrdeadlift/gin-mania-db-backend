# Gin Mania Backend

A starter Go service built with the Gin framework that exposes a world gin search API.

## Prerequisites
- Go 1.22 or newer

## Getting Started
1. Install Go if it is not already available: <https://go.dev/dl/>
2. Pull the Gin dependency and any others:
   ```bash
   go mod tidy
   ```
3. Run the HTTP server:
   ```bash
   go run ./cmd/server
   ```
4. Verify the health endpoint:
   ```bash
   curl http://localhost:8080/healthz
   ```
5. Search for gins using the query parameter `q` (leave empty for all results):
   ```bash
   curl "http://localhost:8080/gins?q=kyoto"
   ```

## Project Layout
- `cmd/server/main.go` – Application entry point and HTTP routes.
- `internal/search` – In-memory search logic and sample catalogue data.

## Next Steps
- Replace the in-memory catalogue with a persistent data store or external API.
- Extend the data model with richer metadata such as ABV, distillery, or tasting notes.
- Add automated tests for the search logic and HTTP handlers.
