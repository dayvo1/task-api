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
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	"strings"

	"github.com/joho/godotenv"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *pgx.Conn

var jwtSecret []byte

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %v\n", r.Method, r.URL.Path, elapsed)
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials Credentials
	json.NewDecoder(r.Body).Decode(&credentials)
	if credentials.Username != os.Getenv("ADMIN_USERNAME") || credentials.Password != os.Getenv("ADMIN_PASSWORD") {
		w.WriteHeader(401)
		return
	}
	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": signed})

}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(401)
			return
		}
		next.ServeHTTP(w, r)

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

	godotenv.Load()

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	r := chi.NewRouter()
	r.Use(loggingMiddleware)
	r.Get("/health", healthHandler)

	r.Post("/login", loginHandler)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/tasks", getTasksHandler)
		r.Get("/tasks/{id}", getTaskHandler)
		r.Post("/tasks", createTaskHandler)
		r.Delete("/tasks/{id}", deleteTaskHandler)

	})

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
