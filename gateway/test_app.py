import json
from unittest.mock import patch, MagicMock
import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as c:
        yield c


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = json.loads(resp.data)
    assert data["status"] == "ok"
    assert data["service"] == "gateway"


def test_create_task_missing_name(client):
    resp = client.post("/api/tasks", json={"description": "no name"})
    assert resp.status_code == 400
    data = json.loads(resp.data)
    assert "error" in data


def test_create_task_no_body(client):
    resp = client.post("/api/tasks", content_type="application/json")
    assert resp.status_code == 400


@patch("app.requests.get")
def test_list_tasks_success(mock_get, client):
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = [{"id": "1", "name": "test"}]
    mock_resp.raise_for_status.return_value = None
    mock_get.return_value = mock_resp

    resp = client.get("/api/tasks")
    assert resp.status_code == 200


@patch("app.requests.get")
def test_list_tasks_engine_down(mock_get, client):
    import requests as req
    mock_get.side_effect = req.ConnectionError("connection refused")

    resp = client.get("/api/tasks")
    assert resp.status_code == 502
    data = json.loads(resp.data)
    assert "unavailable" in data["error"]


@patch("app.requests.post")
def test_create_task_success(mock_post, client):
    mock_resp = MagicMock()
    mock_resp.status_code = 201
    mock_resp.json.return_value = {"id": "1", "name": "my task"}
    mock_resp.raise_for_status.return_value = None
    mock_post.return_value = mock_resp

    resp = client.post("/api/tasks", json={"name": "my task"})
    assert resp.status_code == 201


@patch("app.requests.get")
def test_get_task_not_found(mock_get, client):
    import requests as req
    mock_resp = MagicMock()
    mock_resp.raise_for_status.side_effect = req.exceptions.HTTPError()
    mock_get.return_value = mock_resp

    resp = client.get("/api/tasks/999")
    assert resp.status_code == 404


@patch("app.requests.delete")
def test_delete_task_success(mock_delete, client):
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {"deleted": True}
    mock_resp.raise_for_status.return_value = None
    mock_delete.return_value = mock_resp

    resp = client.delete("/api/tasks/1")
    assert resp.status_code == 200


@patch("app.requests.get")
def test_system_status(mock_get, client):
    mock_resp = MagicMock()
    mock_resp.ok = True
    mock_resp.json.return_value = {"status": "ok"}
    mock_get.return_value = mock_resp

    resp = client.get("/api/status")
    assert resp.status_code == 200
    data = json.loads(resp.data)
    assert "gateway" in data
    assert data["gateway"]["status"] == "ok"
