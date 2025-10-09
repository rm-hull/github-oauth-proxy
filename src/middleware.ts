import { Request, Response, NextFunction } from "express";
import { nanoid } from "nanoid";
import requestIp from "request-ip";
import { logger } from "./logger";
import { httpRequestCounter } from "./metrics";

export const ipMiddleware = requestIp.mw();

export const requestLogger = (req: Request, res: Response, next: NextFunction) => {
  req.log = logger.child({ requestId: nanoid() });
  next();
};

export const metricsMiddleware = (req: Request, res: Response, next: NextFunction) => {
  res.on("finish", () => {
    // Skip metrics endpoint to avoid recursion
    if (req.path === "/metrics" || req.method === "OPTIONS") {
      return;
    }
    // Use req.route.path if available, otherwise fallback to req.path
    // This helps to group metrics by route rather than full path with params
    const route = req.route?.path || "unmatched_route";
    httpRequestCounter.inc({
      method: req.method,
      route: route,
      status_code: res.statusCode,
    });
  });
  next();
};
