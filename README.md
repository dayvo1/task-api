# Task API

A RESTful backend for a task manager built in Go. Includes a Postgres database, JWT authentication, and request logging middleware.

## Tech Stack

- **Go** — HTTP server with the chi router
- **PostgreSQL** — persistent task storage
- **Docker** — containerized app and database
- **JWT** — token-based authentication

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/login` | No | Get a JWT token |
| GET | `/health` | No | Health check |
| GET | `/tasks` | Yes | List all tasks |
| POST | `/tasks` | Yes | Create a task |
| GET | `/tasks/{id}` | Yes | Get a task by ID |
| DELETE | `/tasks/{id}` | Yes | Delete a task |

## Running Locally

**Requirements:** Docker Desktop

```bash
git clone https://github.com/dayvo1/task-api.git
cd task-api
docker compose up --build
```

The server runs on `http://localhost:8080`.

## Usage

**Login to get a token:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

**Use the token on protected routes:**
```bash
curl http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>"
```

**Create a task:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Buy groceries"}'
```

**Delete a task:**
```bash
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <token>"
```
