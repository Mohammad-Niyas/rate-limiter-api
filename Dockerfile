FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (for Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter-api ./cmd/main.go

# --- Production stage ---
FROM alpine:latest

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /app/rate-limiter-api .

EXPOSE 8080

CMD ["./rate-limiter-api"]
