package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// Task API — Step 6: Logging Middleware
//
// Middleware is a function that wraps every request.
// Instead of adding logging to every handler, you write it once here.
//
// How middleware works in Go:
//   Middleware is a function that takes an http.Handler and returns an http.Handler.
//
//   func myMiddleware(next http.Handler) http.Handler {
//       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//           // code here runs BEFORE the handler
//           next.ServeHTTP(w, r)
//           // code here runs AFTER the handler
//       })
//   }
//
//   http.HandlerFunc is just a way to turn a function into an http.Handler.
//   next.ServeHTTP(w, r) calls the actual handler (or next middleware in the chain).
//
// Measuring time:
//   start := time.Now()
//   elapsed := time.Since(start)  // returns a Duration
//
// Logging:
//   log.Printf("%s %s %v\n", r.Method, r.URL.Path, elapsed)
//   %s = string, %v = any value (Duration prints as "1.2ms" etc.)
//
// Registering middleware with chi:
//   r.Use(loggingMiddleware)
//   — must be called before registering routes
//
// Tasks:
//
// 1. Write a loggingMiddleware function:
//    — takes next http.Handler, returns http.Handler
//    — record the start time before calling next
//    — call next.ServeHTTP(w, r)
//    — after it returns, log the method, path, and elapsed time
//
// 2. Register it in main with r.Use(loggingMiddleware)
//    — add it before the route registrations
//
// To run:
//   go run main.go
// Then make any request and watch the terminal — you should see logs like:
//   GET /tasks 1.2ms

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var db *pgx.Conn

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %v\n", r.Method, r.URL.Path, elapsed)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task
	rows, err := db.Query(context.Background(), "SELECT id, title FROM tasks")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Title)
		tasks = append(tasks, t)
	}
	json.NewEncoder(w).Encode(tasks)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var t Task
	err = db.QueryRow(context.Background(), "SELECT id, title FROM tasks WHERE id=$1", id).Scan(&t.ID, &t.Title)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(t)
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var t Task
	json.NewDecoder(r.Body).Decode(&t)
	var id int
	err := db.QueryRow(context.Background(), "INSERT INTO tasks (title) VALUES ($1) RETURNING id", t.Title).Scan(&id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result, err := db.Exec(context.Background(), "DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	r := chi.NewRouter()
	r.Use(loggingMiddleware)
	r.Get("/health", healthHandler)
	r.Get("/tasks", getTasksHandler)
	r.Post("/tasks", createTaskHandler)
	r.Get("/tasks/{id}", getTaskHandler)
	r.Delete("/tasks/{id}", deleteTaskHandler)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
