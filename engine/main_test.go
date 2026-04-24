package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMux() *http.ServeMux {
	store = NewTaskStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/tasks", tasksHandler)
	mux.HandleFunc("/tasks/", taskByIDHandler)
	return mux
}

func TestHealthEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", body["status"])
	}
	if body["service"] != "engine" {
		t.Fatalf("expected service engine, got %s", body["service"])
	}
}

func TestCreateTask(t *testing.T) {
	mux := setupMux()
	payload := `{"name":"test task","description":"a test"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Name != "test task" {
		t.Fatalf("expected name 'test task', got '%s'", task.Name)
	}
	if task.Status != "pending" {
		t.Fatalf("expected status 'pending', got '%s'", task.Status)
	}
}

func TestCreateTaskMissingName(t *testing.T) {
	mux := setupMux()
	payload := `{"description":"no name"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListTasks(t *testing.T) {
	mux := setupMux()

	payload := `{"name":"task1"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/tasks", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tasks []Task
	json.NewDecoder(w.Body).Decode(&tasks)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestGetTask(t *testing.T) {
	mux := setupMux()

	payload := `{"name":"findme"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var created Task
	json.NewDecoder(w.Body).Decode(&created)

	req = httptest.NewRequest("GET", "/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest("GET", "/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	mux := setupMux()

	payload := `{"name":"deleteme"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var created Task
	json.NewDecoder(w.Body).Decode(&created)

	req = httptest.NewRequest("DELETE", "/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest("DELETE", "/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
