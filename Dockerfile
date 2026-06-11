# Step 7: Dockerfile
#
# A Dockerfile tells Docker how to build and run your app as a container.
# We use a multi-stage build — two FROM statements:
#
# Stage 1 (builder): uses the full Go image to compile the binary
# Stage 2 (runner): uses a tiny image to just run the binary
#
# Why two stages? The Go compiler is large (~500MB). The final image only
# needs the compiled binary — no compiler needed. This keeps the image small.
#
# Stage 1 — Build:
#   FROM golang:1.25 AS builder   — start with official Go image, name it "builder"
#   WORKDIR /app                  — set working directory inside the container
#   COPY go.mod go.sum ./         — copy dependency files first (for caching)
#   RUN go mod download           — download dependencies
#   COPY . .                      — copy the rest of the source code
#   RUN go build -o taskapi .     — compile the binary, output named "taskapi"
#
# Stage 2 — Run:
#   FROM debian:bookworm-slim     — start fresh with a tiny base image
#   WORKDIR /app
#   COPY --from=builder /app/taskapi .  — copy only the binary from stage 1
#   EXPOSE 8080                   — document that the app uses port 8080
#   CMD ["./taskapi"]             — command to run when container starts
#
# Tasks:
# 1. Write the Dockerfile using the two stages described above

FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o taskapi

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/taskapi .
EXPOSE 8080
CMD ["./taskapi"]