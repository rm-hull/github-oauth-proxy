import request from "supertest";
import { createServer } from "./server.js";

describe("Server", () => {
  let app: ReturnType<typeof createServer>;

  beforeAll(() => {
    process.env.GITHUB_CLIENT_ID = "test-client-id";
    process.env.GITHUB_CLIENT_SECRET = "test-client-secret";
    process.env.ALLOWED_ORIGINS = "http://localhost:3000";
    process.env.DEFAULT_REDIRECT_URI = "http://localhost:3000/callback";

    app = createServer();
  });

  describe("GET /health", () => {
    it("should return 200 and a message", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body).toEqual({ status: "ok" });
    });
  });

  describe("GET /metrics", () => {
    it("should return 200 and a message", async () => {
      const res = await request(app).get("/metrics");
      expect(res.status).toBe(200);
      expect(res.text).toContain("# HELP");
      expect(res.text).toContain("# TYPE");
    });
  });
});
