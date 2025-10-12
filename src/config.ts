import dotenv from "dotenv";
import { Config, Secret } from "./types/config";

dotenv.config();

const { env } = process;

function extractSecrets(env: NodeJS.ProcessEnv): Record<string, Secret> {
  const result: Record<string, Secret> = {};
  const prefix = "GITHUB_OAUTH_SECRET_";

  for (const [key, value] of Object.entries(env)) {
    if (key.startsWith(prefix) && !!value) {
      const name = key.slice(prefix.length);
      const [clientId, clientSecret] = value.split(/\|/);
      if (!clientId || !clientSecret) {
        throw new Error("Invalid secret format for " + key + ". Expected 'clientId|clientSecret'.");
      }
      result[clientId] = { name, clientId, clientSecret };
    }
  }

  return result;
}

export const config: Config = {
  port: Number(env.PORT) || 3001,
  logLevel: env.LOG_LEVEL || "info",
  nodeEnv: env.NODE_ENV || "development",
  github: extractSecrets(env),
  cors: {
    allowedOrigins: (env.ALLOWED_ORIGINS || "http://localhost:5173").split(","),
  },
};

if (!config.github || Object.keys(config.github).length === 0) {
  throw new Error("Missing GitHub OAuth credentials");
}
