import pino from "pino";
import { config } from "./config";

export const logger = pino({
  level: config.logLevel,
  formatters: {
    level: (label) => ({ level: label.toUpperCase() }),
  },
  transport:
    config.nodeEnv === "development"
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
