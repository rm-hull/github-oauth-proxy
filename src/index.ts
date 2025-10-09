import express from "express";
import client from "prom-client";
import cors from "cors";
import dotenv from "dotenv";
import pino from "pino";
import { nanoid } from "nanoid";
import requestIp from "request-ip";

dotenv.config();

declare module "express-serve-static-core" {
  interface Request {
    log: pino.Logger;
  }
}

// Initialize logger
const logger = pino({
  level: process.env.LOG_LEVEL || "info",
  formatters: {
    level: (label) => ({ level: label.toUpperCase() }),
  },
  transport:
    process.env.NODE_ENV === "development"
      ? {
          target: "pino-pretty",
          options: {
            colorize: true,
            translateTime: "HH:MM:ss",
            ignore: "pid,hostname",
          },
        }
      : undefined,
});

// Types for GitHub OAuth responses
interface GitHubTokenResponse {
  access_token?: string;
  token_type?: string;
  scope?: string;
  error?: string;
  error_description?: string;
  error_uri?: string;
}

const app = express();
app.disable("x-powered-by");

const PORT = process.env.PORT || 3001;
const CLIENT_ID = process.env.GITHUB_CLIENT_ID;
const CLIENT_SECRET = process.env.GITHUB_CLIENT_SECRET;

if (!CLIENT_ID || !CLIENT_SECRET) {
  logger.error(
    { clientId: !!CLIENT_ID, clientSecret: !!CLIENT_SECRET },
    "Missing GitHub OAuth credentials"
  );

  process.exit(1);
}

// Prometheus metrics setup
const collectDefaultMetrics = client.collectDefaultMetrics;
collectDefaultMetrics();

const httpRequestCounter = new client.Counter({
  name: "http_requests_total",
  help: "Total number of HTTP requests",
  labelNames: ["method", "route", "status_code"],
});

app.use((req, res, next) => {
  res.on("finish", () => {
    if (req.path === "/metrics" || req.method === "OPTIONS") return; // Skip metrics endpoint to avoid recursion
    httpRequestCounter.inc({
      method: req.method,
      route: req.path,
      status_code: res.statusCode,
    });
  });
  next();
});

app.use(requestIp.mw());

app.use((req, res, next) => {
  req.log = logger.child({ requestId: nanoid() });
  next();
});

app.use(
  cors({
    origin: process.env.ALLOWED_ORIGINS?.split(",") || "http://localhost:5173",
    credentials: true,
  })
);

app.use(express.json());

// Metrics endpoint
app.get("/metrics", async (req, res) => {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
});

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({ status: "ok" });
});

// GitHub OAuth token exchange endpoint
app.post("/v1/github/token", async (req, res) => {
  const { code, code_verifier, redirect_uri } = req.body;

  if (!code) {
    return res.status(400).json({ error: "Missing code parameter" });
  }

  if (!code_verifier) {
    return res.status(400).json({ error: "Missing code_verifier parameter" });
  }

  if (!redirect_uri) {
    return res.status(400).json({ error: "Missing redirect_uri parameter" });
  }

  const allowedOrigins = process.env.ALLOWED_ORIGINS?.split(",") || [];
  try {
    const redirectUriOrigin = new URL(redirect_uri).origin;
    if (!allowedOrigins.includes(redirectUriOrigin)) {
      req.log.warn({ redirect_uri }, "Invalid redirect_uri received.");
      return res.status(400).json({ error: "Invalid redirect_uri" });
    }
  } catch (e) {
    req.log.warn({ redirect_uri }, "Invalid redirect_uri format received.");
    return res.status(400).json({ error: "Invalid redirect_uri format" });
  }

  try {
    req.log.info(
      {
        clientId: CLIENT_ID,
        redirectUri: redirect_uri,
        hasCode: !!code,
        hasVerifier: !!code_verifier,
        ip: req.clientIp,
      },
      "Exchanging code for token"
    );

    const response = await fetch(
      "https://github.com/login/oauth/access_token",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({
          client_id: CLIENT_ID,
          client_secret: CLIENT_SECRET,
          code,
          code_verifier,
          redirect_uri: redirect_uri,
        }),
      }
    );

    const data = (await response.json()) as GitHubTokenResponse;

    // Don't expose client credentials in error responses
    if (data.error) {
      req.log.error(
        {
          error: data.error,
          description: data.error_description,
        },
        "GitHub OAuth error"
      );
      return res.status(400).json({
        error: data.error,
        error_description: data.error_description,
      });
    }

    req.log.info(
      {
        hasAccessToken: !!data.access_token,
        scope: data.scope,
        tokenType: data.token_type,
      },
      "Token exchange successful"
    );
    res.json(data);
  } catch (error) {
    req.log.error({ error }, "Error exchanging code for token");
    res.status(500).json({ error: "Internal server error" });
  }
});

app
  .listen(PORT, () => {
    logger.info({ port: PORT }, "GitHub OAuth proxy server started");
  })
  .on("error", (err) => {
    if ("code" in err && err.code === "EADDRINUSE") {
      logger.error({ port: PORT }, "Port is already in use.");
    } else {
      logger.error({ err }, "Failed to start server");
    }
    process.exit(1);
  });
