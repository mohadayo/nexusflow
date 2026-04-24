import express, { Request, Response } from "express";
import cors from "cors";
import axios from "axios";
import dotenv from "dotenv";

dotenv.config();

const app = express();
app.use(cors());
app.use(express.json());

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8080";
const PORT = parseInt(process.env.DASHBOARD_PORT || "8082", 10);

app.get("/health", (_req: Request, res: Response) => {
  res.json({ status: "ok", service: "dashboard" });
});

app.get("/dashboard/summary", async (_req: Request, res: Response) => {
  console.log(`[dashboard] Fetching task summary from gateway`);
  try {
    const response = await axios.get(`${GATEWAY_URL}/api/tasks`, {
      timeout: 5000,
    });
    const tasks = response.data as Array<{ status: string }>;
    const summary = {
      total: tasks.length,
      pending: tasks.filter((t) => t.status === "pending").length,
      completed: tasks.filter((t) => t.status === "completed").length,
      in_progress: tasks.filter((t) => t.status === "in_progress").length,
    };
    res.json(summary);
  } catch (error) {
    console.error("[dashboard] Failed to fetch tasks:", error);
    res.status(502).json({ error: "Gateway service unavailable" });
  }
});

app.get("/dashboard/tasks", async (_req: Request, res: Response) => {
  console.log(`[dashboard] Proxying task list from gateway`);
  try {
    const response = await axios.get(`${GATEWAY_URL}/api/tasks`, {
      timeout: 5000,
    });
    res.json(response.data);
  } catch (error) {
    console.error("[dashboard] Failed to fetch tasks:", error);
    res.status(502).json({ error: "Gateway service unavailable" });
  }
});

app.get("/dashboard/status", async (_req: Request, res: Response) => {
  console.log(`[dashboard] Fetching system status`);
  try {
    const response = await axios.get(`${GATEWAY_URL}/api/status`, {
      timeout: 5000,
    });
    res.json(response.data);
  } catch (error) {
    console.error("[dashboard] Failed to fetch status:", error);
    res.status(502).json({ error: "Gateway service unavailable" });
  }
});

export { app };

if (require.main === module) {
  app.listen(PORT, "0.0.0.0", () => {
    console.log(`[dashboard] Starting dashboard BFF on port ${PORT}`);
  });
}
