# GitHub OAuth Proxy

## Summary

GitHub OAuth Proxy is a lightweight Node.js service that acts as a secure intermediary between a browser-based single page application and GitHub's OAuth authentication. It is designed specifically to support OAuth flows using PKCE (Proof Key for Code Exchange), making it ideal for SPAs and mobile apps. The proxy simplifies OAuth logic, handles token exchanges, and provides endpoints for authenticating users via GitHub, so you can integrate GitHub login without exposing secrets or handling complex OAuth and PKCE logic directly.

## Usage Examples

### Authenticate a user

```http
POST http://localhost:3001/v1/auth/github
Content-Type: application/json

{
  "code": "<github-oauth-code>",
  "code_verifier": "<github-pkce-verifier>"
}
```

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
