import os
import logging
import requests
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("gateway")

ENGINE_URL = os.getenv("ENGINE_URL", "http://localhost:8081")
DASHBOARD_URL = os.getenv("DASHBOARD_URL", "http://localhost:8082")


@app.route("/health")
def health():
    return jsonify({"status": "ok", "service": "gateway"})


@app.route("/api/tasks", methods=["GET"])
def list_tasks():
    logger.info("Listing tasks via engine")
    try:
        resp = requests.get(f"{ENGINE_URL}/tasks", timeout=5)
        resp.raise_for_status()
        return jsonify(resp.json()), resp.status_code
    except requests.RequestException as e:
        logger.error("Failed to reach engine: %s", e)
        return jsonify({"error": "Engine service unavailable"}), 502


@app.route("/api/tasks", methods=["POST"])
def create_task():
    body = request.get_json()
    if not body or "name" not in body:
        return jsonify({"error": "Field 'name' is required"}), 400
    logger.info("Creating task: %s", body.get("name"))
    try:
        resp = requests.post(f"{ENGINE_URL}/tasks", json=body, timeout=5)
        resp.raise_for_status()
        return jsonify(resp.json()), resp.status_code
    except requests.RequestException as e:
        logger.error("Failed to reach engine: %s", e)
        return jsonify({"error": "Engine service unavailable"}), 502


@app.route("/api/tasks/<task_id>", methods=["GET"])
def get_task(task_id):
    logger.info("Getting task: %s", task_id)
    try:
        resp = requests.get(f"{ENGINE_URL}/tasks/{task_id}", timeout=5)
        resp.raise_for_status()
        return jsonify(resp.json()), resp.status_code
    except requests.exceptions.HTTPError:
        return jsonify({"error": "Task not found"}), 404
    except requests.RequestException as e:
        logger.error("Failed to reach engine: %s", e)
        return jsonify({"error": "Engine service unavailable"}), 502


@app.route("/api/tasks/<task_id>", methods=["DELETE"])
def delete_task(task_id):
    logger.info("Deleting task: %s", task_id)
    try:
        resp = requests.delete(f"{ENGINE_URL}/tasks/{task_id}", timeout=5)
        resp.raise_for_status()
        return jsonify(resp.json()), resp.status_code
    except requests.exceptions.HTTPError:
        return jsonify({"error": "Task not found"}), 404
    except requests.RequestException as e:
        logger.error("Failed to reach engine: %s", e)
        return jsonify({"error": "Engine service unavailable"}), 502


@app.route("/api/status")
def system_status():
    logger.info("Checking system status")
    statuses = {}
    for name, url in [("engine", ENGINE_URL), ("dashboard", DASHBOARD_URL)]:
        try:
            resp = requests.get(f"{url}/health", timeout=3)
            statuses[name] = resp.json() if resp.ok else {"status": "unhealthy"}
        except requests.RequestException:
            statuses[name] = {"status": "unreachable"}
    statuses["gateway"] = {"status": "ok", "service": "gateway"}
    return jsonify(statuses)


if __name__ == "__main__":
    port = int(os.getenv("GATEWAY_PORT", "8080"))
    logger.info("Starting gateway on port %d", port)
    app.run(host="0.0.0.0", port=port)
