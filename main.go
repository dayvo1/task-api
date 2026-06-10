package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// Task API — Step 5: Postgres Database
//
// Instead of storing tasks in a slice, we'll store them in Postgres.
// Data now persists across server restarts.
//
// Connecting to Postgres:
//   conn, err := pgx.Connect(context.Background(), "postgres://user:password@host:port/dbname")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer conn.Close(context.Background())
//
//   Connection string for our docker setup:
//   "postgres://postgres:password@localhost:5432/taskapi"
//
// Running a query that returns rows:
//   rows, err := conn.Query(ctx, "SELECT id, title FROM tasks")
//   for rows.Next() {
//       var t Task
//       rows.Scan(&t.ID, &t.Title)
//       tasks = append(tasks, t)
//   }
//
// Running a query that returns one row:
//   var t Task
//   err := conn.QueryRow(ctx, "SELECT id, title FROM tasks WHERE id=$1", id).Scan(&t.ID, &t.Title)
//   if err == pgx.ErrNoRows {
//       // not found
//   }
//
// INSERT and get the new ID back:
//   var id int
//   err := conn.QueryRow(ctx, "INSERT INTO tasks (title) VALUES ($1) RETURNING id", t.Title).Scan(&id)
//
// DELETE and check if anything was deleted:
//   result, err := conn.Exec(ctx, "DELETE FROM tasks WHERE id=$1", id)
//   if result.RowsAffected() == 0 {
//       // not found
//   }
//
// context.Background() is just a default context — ignore it for now,
// it's required by the pgx API.
//
// Tasks:
//
// 1. Declare a package-level variable: var db *pgx.Conn
//
// 2. In main, connect to Postgres before starting the server:
//    — use the connection string for our docker setup
//    — store the result in db
//    — log.Fatal if it fails
//    — defer db.Close(context.Background())
//
// 3. Update getTasksHandler:
//    — query all rows from the tasks table
//    — scan each row into a Task and append to a local slice
//    — encode and return the slice
//
// 4. Update getTaskHandler:
//    — use QueryRow with WHERE id=$1
//    — if pgx.ErrNoRows, return 404
//    — otherwise encode the task
//
// 5. Update createTaskHandler:
//    — INSERT the title into the database
//    — use RETURNING id to get the new ID back
//    — respond with 201 and the new task
//
// 6. Update deleteTaskHandler:
//    — DELETE from the database WHERE id=$1
//    — check RowsAffected() — if 0, return 404
//    — otherwise return 204

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var db *pgx.Conn

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
	w.WriteHeader(http.StatusOK)
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
		w.WriteHeader(500)
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
		w.WriteHeader(500)
		return
	}
	if result.RowsAffected() == 0 {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {

	var err error
	db, err = pgx.Connect(context.Background(), "postgres://postgres:password@localhost:5432/taskapi")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	r := chi.NewRouter()
	r.Get("/health", healthHandler)
	r.Get("/tasks", getTasksHandler)
	r.Post("/tasks", createTaskHandler)
	r.Get("/tasks/{id}", getTaskHandler)
	r.Delete("/tasks/{id}", deleteTaskHandler)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
