# GEMINI.md

## Project Overview

This project is a lightweight Go service that acts as a secure intermediary between a browser-based single page application and GitHub's OAuth authentication. It is designed specifically to support OAuth flows using PKCE (Proof Key for Code Exchange), making it ideal for SPAs and mobile apps. The proxy simplifies OAuth logic, handles token exchanges, and provides endpoints for authenticating users via GitHub, so you can integrate GitHub login without exposing secrets or handling complex OAuth and PKCE logic directly.

The application is built with Go and uses the Gin framework. It follows the architectural patterns found in https://github.com/map-services/fuel-prices-api.

## Building and Running

### Prerequisites

- Go 1.26 or later

### Configuration

Create a `.env` file from the example and add your GitHub OAuth credentials:

```bash
cp .env.example .env
```

### Development

To run the application:

```bash
go run main.go
```

### Production

To build and run the application:

```bash
go build -o github-oauth-proxy .
./github-oauth-proxy
```

### Docker

To build and run the application using Docker:

```bash
docker build -t github-oauth-proxy .
docker run -p 8080:8080 --env-file .env github-oauth-proxy
```

### Testing

To run the tests:

```bash
go test -v ./...
```

To check linting:

```bash
golangci-lint-v2 run ./...
```

## Development Conventions

*   **Language:** Go
*   **Framework:** Gin Gonic
*   **CLI:** Cobra
*   **Logging:** log/slog
*   **Metrics:** Prometheus (via ginprom)
*   **Health Checks:** gin-healthcheck
*   **Internal logic:** All core logic resides in the `internal/` directory.
