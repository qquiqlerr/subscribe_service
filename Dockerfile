# Build stage
FROM golang:1.23.8-alpine AS builder

WORKDIR /app

# Copy go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN go build -o server ./cmd/main.go

# Production stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Copy config file
COPY config.yaml .

# Start the server
ENTRYPOINT ["./server"]
