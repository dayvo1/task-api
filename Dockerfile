# start with official Go image, name this stage "builder"
FROM golang:1.25 AS builder
# set working directory inside the container
WORKDIR /app
# copy dependency files first (for layer caching)
COPY go.mod go.sum ./
# download dependencies
RUN go mod download
# copy the rest of the source code
COPY . .
# compile the binary, output named "taskapi"
RUN go build -o taskapi .

# start fresh with a tiny base image (no compiler)
FROM debian:bookworm-slim
WORKDIR /app
# copy only the binary from the builder stage
COPY --from=builder /app/taskapi .
# document that the app listens on port 8080
EXPOSE 8080
# command to run when the container starts
CMD ["./taskapi"]
