# Build stage
FROM golang:1.21.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git tree

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Show directory structure and file contents at each step
RUN echo "=== Initial directory structure ===" && \
  tree && \
  echo "=== Initial template files ===" && \
  find . -name "*.templ" -type f -exec sh -c 'echo "=== Contents of {} ==="; cat {}' \;

# Generate templ templates
RUN echo "=== Generating templates ===" && \
  templ generate ./internal/templates && \
  echo "=== After generation ==="

# Show generated files
RUN echo "=== Directory structure after generation ===" && \
  tree && \
  echo "=== Generated template files ===" && \
  find . -name "*_templ.go" -type f -exec sh -c 'echo "=== Contents of {} ==="; cat {}' \;

# Try to compile just the templates package first
RUN echo "=== Compiling templates package ===" && \
  cd internal/templates && \
  go list -f '{{.GoFiles}}' && \
  cd ../..

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/server ./cmd/server

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Run the application
CMD ["./server"]
