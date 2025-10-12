import { Router } from "express";
import client from "prom-client";
import { config } from "./config";
import { GitHubTokenResponse } from "./types/github";

export const router = Router();

// Metrics endpoint
router.get("/metrics", async (req, res) => {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
});

// Health check endpoint
router.get("/health", (req, res) => {
  res.json({ status: "ok" });
});

// GitHub OAuth token exchange endpoint
router.post("/v1/github/token", async (req, res) => {
  const { client_id, code, code_verifier, redirect_uri } = req.body;

  const requiredParams = ["client_id", "code", "code_verifier", "redirect_uri"];
  for (const param of requiredParams) {
    if (!(param in req.body)) {
      req.log.warn({ param }, `Missing parameter`);
      return res.status(400).json({ error: `Missing ${param} parameter` });
    }
  }

  const allowedOrigins = config.cors.allowedOrigins;
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

  const secret = config.github[client_id];
  if (!secret) {
    req.log.warn({ client_id }, "Unknown client_id received.");
    return res.status(400).json({ error: "Invalid client_id" });
  }

  try {
    req.log.info(
      {
        app: secret.name,
        clientId: client_id,
        redirectUri: redirect_uri,
        hasCode: !!code,
        hasVerifier: !!code_verifier,
        ip: req.clientIp,
      },
      "Exchanging code for token"
    );

    const response = await fetch("https://github.com/login/oauth/access_token", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({
        client_id: secret.clientId,
        client_secret: secret.clientSecret,
        code,
        code_verifier,
        redirect_uri,
      }),
    });

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
