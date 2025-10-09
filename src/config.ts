import dotenv from "dotenv";

dotenv.config();

const { env } = process;

export const config = {
  port: Number(env.PORT) || 3001,
  logLevel: env.LOG_LEVEL || "info",
  nodeEnv: env.NODE_ENV || "development",
  github: {
    clientId: env.GITHUB_CLIENT_ID,
    clientSecret: env.GITHUB_CLIENT_SECRET,
  },
  cors: {
    allowedOrigins: env.ALLOWED_ORIGINS?.split(",") || [
      "http://localhost:5173",
    ],
  },
};

if (!config.github.clientId || !config.github.clientSecret) {
  throw new Error("Missing GitHub OAuth credentials");
}
