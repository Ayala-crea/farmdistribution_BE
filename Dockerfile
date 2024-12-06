# Stage 1: Build
FROM golang:1.22 AS builder

# Set work directory
WORKDIR /app

# Copy go.mod and go.sum first (to leverage Docker layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o main .

# Stage 2: Run
FROM debian:bullseye-slim

# Set work directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on
EXPOSE 8080

# Run the binary
CMD ["/app/main"]
