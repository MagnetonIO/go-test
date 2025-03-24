# Use official Golang image
FROM golang:1.20

# Set the working directory
WORKDIR /app

# Copy the Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o app

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./app"]
