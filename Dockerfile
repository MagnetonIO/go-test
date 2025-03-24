FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first to leverage Docker cache
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o btc-ltp-service .

# Use a smaller image for the final build
FROM alpine:latest

WORKDIR /app

# Add ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/btc-ltp-service .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./btc-ltp-service"]
