import { createServer } from "./server";
import { config } from "./config";
import { logger } from "./logger";

const app = createServer();

app
  .listen(config.port, () => {
    logger.info({ port: config.port }, "GitHub OAuth proxy server started");
  })
  .on("error", (err) => {
    if ("code" in err && err.code === "EADDRINUSE") {
      logger.error({ port: config.port }, "Port is already in use.");
    } else {
      logger.error({ err }, "Failed to start server");
    }
    process.exit(1);
  });