.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o btc-ltp-service

# Run the application
run: build
	./btc-ltp-service

# Run tests
test:
	go test -v

# Run tests without integration tests
test-short:
	go test -v -short

# Clean build artifacts
clean:
	rm -f btc-ltp-service

# Build Docker image
docker-build:
	docker build -t btc-ltp-service .

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 btc-ltp-service

# Run with Docker Compose
docker-compose-up:
	docker-compose up --build

# Stop Docker Compose services
docker-compose-down:
	docker-compose down
