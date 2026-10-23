package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Tasks struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Subject  string `json:"subject"`
	Status   string `json:"status"`
	Deadline string `json:"deadline"`
}

type UpdateTaskRequest struct {
	Title    *string `json:"title"`
	Subject  *string `json:"subject"`
	Status   *string `json:"status"`
	Deadline *string `json:"deadline"`
}

var (
	tasks  = make(map[int]Tasks)
	mu     sync.Mutex
	nextID = 1
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", getTasks)
	mux.HandleFunc("POST /tasks", createTasks)
	mux.HandleFunc("PATCH /tasks/{id}", updateTasks)
	mux.HandleFunc("DELETE /tasks/{id}", deleteTasks)

	server := &http.Server{
		Addr:              ":8082",
		Handler:           loggingMiddleware(mux),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// get tasks
// post /tasks
// patch /tasks/{id}
// delete /tasks/{id}
// json

func getTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	list := make([]Tasks, 0, len(tasks))
	for _, task := range tasks {
		list = append(list, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func createTasks(w http.ResponseWriter, r *http.Request) {
	var req Tasks
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	req.ID = nextID
	nextID++
	tasks[req.ID] = req
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func updateTasks(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	task, exists := tasks[id]
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Subject != nil {
		task.Subject = *req.Subject
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Deadline != nil {
		task.Deadline = *req.Deadline
	}

	tasks[id] = task

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func deleteTasks(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := tasks[id]; !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	delete(tasks, id)
	w.WriteHeader(http.StatusNoContent)
}
