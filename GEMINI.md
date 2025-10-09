# GEMINI.md

## Project Overview

This project is a lightweight Node.js service that acts as a secure intermediary between a browser-based single page application and GitHub's OAuth authentication. It is designed specifically to support OAuth flows using PKCE (Proof Key for Code Exchange), making it ideal for SPAs and mobile apps. The proxy simplifies OAuth logic, handles token exchanges, and provides endpoints for authenticating users via GitHub, so you can integrate GitHub login without exposing secrets or handling complex OAuth and PKCE logic directly.

The application is built with TypeScript and uses Express.js for the web server. It also includes `pino` for logging and `prom-client` for exposing Prometheus metrics.

## Building and Running

### Installation

```bash
yarn install
```

### Configuration

Create a `.env` file from the example and add your GitHub OAuth credentials:

```bash
cp .env.example .env
```

### Development

To run the application in development mode with hot-reloading:

```bash
yarn dev
```

### Production

To build and run the application in production mode:

```bash
yarn build
yarn start
```

### Docker

To build and run the application using Docker:

```bash
docker build -t github-oauth-proxy .
docker run -p 3001:3001 --env-file .env github-oauth-proxy
```

### Type Checking

To check for TypeScript type errors:

```bash
yarn type-check
```

## Development Conventions

*   **Language:** TypeScript
*   **Framework:** Express.js
*   **Package Manager:** Yarn
*   **Logging:** Pino is used for logging. In development, `pino-pretty` is used for human-readable logs.
*   **Metrics:** Prometheus metrics are exposed on the `/metrics` endpoint using `prom-client`.
*   **Code Style:** The project uses Prettier for code formatting (inferred from the presence of `.prettierrc`).
*   **Linting:** ESLint is used for linting (inferred from the presence of `.eslintrc.js`).
