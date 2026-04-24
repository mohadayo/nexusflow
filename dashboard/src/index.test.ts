import request from "supertest";
import axios from "axios";
import { app } from "./index";

jest.mock("axios");
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe("Dashboard BFF", () => {
  describe("GET /health", () => {
    it("returns ok status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("ok");
      expect(res.body.service).toBe("dashboard");
    });
  });

  describe("GET /dashboard/summary", () => {
    it("returns task summary", async () => {
      mockedAxios.get.mockResolvedValueOnce({
        data: [
          { id: "1", name: "t1", status: "pending" },
          { id: "2", name: "t2", status: "completed" },
          { id: "3", name: "t3", status: "pending" },
        ],
      });

      const res = await request(app).get("/dashboard/summary");
      expect(res.status).toBe(200);
      expect(res.body.total).toBe(3);
      expect(res.body.pending).toBe(2);
      expect(res.body.completed).toBe(1);
    });

    it("returns 502 when gateway is down", async () => {
      mockedAxios.get.mockRejectedValueOnce(new Error("ECONNREFUSED"));

      const res = await request(app).get("/dashboard/summary");
      expect(res.status).toBe(502);
      expect(res.body.error).toContain("unavailable");
    });
  });

  describe("GET /dashboard/tasks", () => {
    it("proxies task list", async () => {
      mockedAxios.get.mockResolvedValueOnce({
        data: [{ id: "1", name: "task1" }],
      });

      const res = await request(app).get("/dashboard/tasks");
      expect(res.status).toBe(200);
      expect(res.body).toHaveLength(1);
    });
  });

  describe("GET /dashboard/status", () => {
    it("proxies system status", async () => {
      mockedAxios.get.mockResolvedValueOnce({
        data: {
          gateway: { status: "ok" },
          engine: { status: "ok" },
        },
      });

      const res = await request(app).get("/dashboard/status");
      expect(res.status).toBe(200);
      expect(res.body.gateway.status).toBe("ok");
    });

    it("returns 502 when gateway is down", async () => {
      mockedAxios.get.mockRejectedValueOnce(new Error("ECONNREFUSED"));

      const res = await request(app).get("/dashboard/status");
      expect(res.status).toBe(502);
    });
  });
});
