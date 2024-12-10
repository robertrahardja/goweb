FROM golang:1.21-alpine as builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.* .
RUN go mod download && \
  go mod verify

# Copy the rest of the source code
COPY . .

# Install templ and generate templates
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Create final minimal image
FROM alpine:latest

# Add CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D appuser

WORKDIR /app

# Copy binary and certificates from builder
COPY --from=builder /app/main .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user
USER appuser

EXPOSE 3000

ENTRYPOINT ["./main"]
