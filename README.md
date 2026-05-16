# GitHub OAuth Proxy

## Summary

GitHub OAuth Proxy is a lightweight Go service that acts as a secure intermediary between a browser-based single page application and GitHub's OAuth authentication. It is designed specifically to support OAuth flows using PKCE (Proof Key for Code Exchange), making it ideal for SPAs and mobile apps. The proxy simplifies OAuth logic, handles token exchanges, and provides endpoints for authenticating users via GitHub, so you can integrate GitHub login without exposing secrets or handling complex OAuth and PKCE logic directly.

## Usage Examples

### Authenticate a user

```http
POST http://localhost:8080/v1/github/token
Content-Type: application/json

{
  "client_id": "<github-client-id>",
  "code": "<github-oauth-code>",
  "code_verifier": "<github-pkce-verifier>",
  "redirect_uri": "<github-redirect-uri>"
}
```

The redirect URI must match the value specified in the GitHub application and must be one of the `ALLOWED_ORIGINS`.

## Local setup

```console
# Create .env file from example
cp .env.example .env
# Edit .env with your actual credentials

# Run development
go run main.go

# Build
go build -o github-oauth-proxy .

# Run tests
go test -v ./...

# Docker
docker build -t github-oauth-proxy .
docker run -p 8080:8080 --env-file .env github-oauth-proxy
```

## Configuration

### Environment Variables

The following environment variables are used to configure the application:

| Variable                     | Description                                                                                                                                    | Default                 |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `PORT`                       | The port to run the application on.                                                                                                            | `8080`                  |
| `LOG_LEVEL`                  | The log level to use (debug, info, warn, error).                                                                                               | `info`                  |
| `GITHUB_OAUTH_SECRET_<NAME>` | A GitHub OAuth secret, where `<NAME>` is a unique identifier for the client. The value should be in the format `<client_id>\|<client_secret>`. |                         |
| `ALLOWED_ORIGINS`            | A comma-separated list of allowed CORS origins.                                                                                                | `http://localhost:5173` |

## Logging

The application uses `zerolog` for structured logging. In `debug` mode, logs are pretty-printed to the console. In other modes, they are formatted as JSON.

## Metrics

The application exposes Prometheus metrics on the `/metrics` endpoint.

## Health Check

The application provides a health check on `/healthz` (provided by `gin-healthcheck`).

