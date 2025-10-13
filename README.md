# GitHub OAuth Proxy

## Summary

GitHub OAuth Proxy is a lightweight Node.js service that acts as a secure intermediary between a browser-based single page application and GitHub's OAuth authentication. It is designed specifically to support OAuth flows using PKCE (Proof Key for Code Exchange), making it ideal for SPAs and mobile apps. The proxy simplifies OAuth logic, handles token exchanges, and provides endpoints for authenticating users via GitHub, so you can integrate GitHub login without exposing secrets or handling complex OAuth and PKCE logic directly.

## Usage Examples

### Authenticate a user

```http
POST http://localhost:3001/v1/auth/github
Content-Type: application/json

{
  "client_id": "<github-client-id>",
  "code": "<github-oauth-code>",
  "code_verifier": "<github-pkce-verifier>",
  "redirect_uri": "<github-redirect-uri>"   
}
```
The redirect URI must match the value specified in the GitHub application. 

## Local setup

```console
# Install dependencies
yarn install

# Create .env file from example
cp .env.example .env
# Edit .env with your actual credentials

# Development
yarn dev

# Build and run production
yarn build
yarn start

# Docker
docker build -t github-oauth-proxy .
docker run -p 3001:3001 --env-file .env github-oauth-proxy
```

## Configuration

### Environment Variables

The following environment variables are used to configure the application:

| Variable                     | Description                                                                                                                                    | Default                 |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `PORT`                       | The port to run the application on.                                                                                                            | `3001`                  |
| `LOG_LEVEL`                  | The log level to use.                                                                                                                          | `info`                  |
| `GITHUB_OAUTH_SECRET_<NAME>` | A GitHub OAuth secret, where `<NAME>` is a unique identifier for the client. The value should be in the format `<client_id>\|<client_secret>`. |                         |
| `ALLOWED_ORIGINS`            | A comma-separated list of allowed CORS origins.                                                                                                | `http://localhost:5173` |

## Logging

The application uses Pino for logging. In development, the logs are pretty-printed to the console. In production, the logs are formatted as JSON.

Each log entry will contain a `requestId` that is unique to each request. This can be used to correlate all the logs for a single request.

The log levels are as follows:

- `FATAL`
- `ERROR`
- `WARN`
- `INFO`
- `DEBUG`
- `TRACE`

## Metrics

The application exposes Prometheus metrics on the `/metrics` endpoint. The following metrics are available:

- `http_requests_total`: The total number of HTTP requests, with labels for `method`, `route`, and `status_code`.
