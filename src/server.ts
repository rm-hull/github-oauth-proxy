import cors from "cors";
import express from "express";
import { config } from "./config";
import { collectDefaultMetrics } from "./metrics";
import { ipMiddleware, requestLogger, metricsMiddleware } from "./middleware";
import { router } from "./routes";

declare module "express-serve-static-core" {
  interface Request {
    log: import("pino").Logger;
  }
}

export const createServer = () => {
  const app = express();
  app.disable("x-powered-by");

  collectDefaultMetrics();

  app.use(ipMiddleware);
  app.use(requestLogger);
  app.use(metricsMiddleware);
  app.use(cors({ origin: config.cors.allowedOrigins, credentials: true }));
  app.use(express.json());
  app.use(router);

  return app;
};
