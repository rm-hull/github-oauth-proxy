// src/index.ts
import express from "express";
import cors from "cors";
import dotenv from "dotenv";
import pino from "pino";

dotenv.config();

// Initialize logger
const logger = pino({
  level: process.env.LOG_LEVEL || "info",
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
const PORT = process.env.PORT || 3001;

// Configure CORS - adjust origins based on your needs
app.use(
  cors({
    origin: process.env.ALLOWED_ORIGINS?.split(",") || "http://localhost:5173",
    credentials: true,
  })
);

app.use(express.json());

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({ status: "ok" });
});

// GitHub OAuth token exchange endpoint
app.post("/v1/github/token", async (req, res) => {
  const { code, code_verifier } = req.body;

  if (!code) {
    return res.status(400).json({ error: "Missing code parameter" });
  }

  const clientId = process.env.GITHUB_CLIENT_ID;
  const clientSecret = process.env.GITHUB_CLIENT_SECRET;
  const redirectUri = process.env.REDIRECT_URI;

  if (!clientId || !clientSecret) {
    logger.error(
      { clientId: !!clientId, clientSecret: !!clientSecret },
      "Missing GitHub OAuth credentials"
    );
    return res.status(500).json({ error: "Server configuration error" });
  }

  try {
    logger.info(
      { clientId, hasCode: !!code, hasVerifier: !!code_verifier },
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
          client_id: clientId,
          client_secret: clientSecret,
          code,
          code_verifier,
          redirect_uri: redirectUri,
        }),
      }
    );

    const data = (await response.json()) as GitHubTokenResponse;

    // Don't expose client credentials in error responses
    if (data.error) {
      logger.error(
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

    logger.info(
      { hasAccessToken: !!data.access_token, scope: data.scope },
      "Token exchange successful"
    );
    res.json(data);
  } catch (error) {
    logger.error({ error }, "Error exchanging code for token");
    res.status(500).json({ error: "Internal server error" });
  }
});

app.listen(PORT, () => {
  logger.info({ port: PORT }, "GitHub OAuth proxy server started");
});
