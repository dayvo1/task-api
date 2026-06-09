package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var tasks = []Task{
	{ID: 1, Title: "Buy groceries"},
	{ID: 2, Title: "Walk the dog"},
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		w.Write([]byte(`{"message":"task received"}`))
	} else if r.Method == "GET" {
		json.NewEncoder(w).Encode(tasks)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/tasks", tasksHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
