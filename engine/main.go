package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]Task)}
}

func (s *TaskStore) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *TaskStore) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *TaskStore) Create(name, description string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	s.tasks[t.ID] = t
	return t
}

func (s *TaskStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

var store = NewTaskStore()
var logger = log.New(os.Stdout, "[engine] ", log.LstdFlags)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "engine"})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		logger.Println("Listing all tasks")
		writeJSON(w, http.StatusOK, store.List())
	case http.MethodPost:
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		if body.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Field 'name' is required"})
			return
		}
		logger.Printf("Creating task: %s\n", body.Name)
		task := store.Create(body.Name, body.Description)
		writeJSON(w, http.StatusCreated, task)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tasks/"):]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Task ID required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		logger.Printf("Getting task: %s\n", id)
		task, ok := store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
			return
		}
		writeJSON(w, http.StatusOK, task)
	case http.MethodDelete:
		logger.Printf("Deleting task: %s\n", id)
		if !store.Delete(id) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("ENGINE_PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/tasks", tasksHandler)
	mux.HandleFunc("/tasks/", taskByIDHandler)

	logger.Printf("Starting engine on port %s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("Server failed: %v\n", err)
	}
}
