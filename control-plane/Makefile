.PHONY: build run clean test redis-start redis-stop

# Binary name
BINARY_NAME=skyscale-control-plane

# Redis configuration
REDIS_CONTAINER=faas-redis

# Build the control plane
build:
	go build -o $(BINARY_NAME) .

# Start Redis container
redis-start:
	@echo "Starting Redis container..."
	@docker ps -q -f name=$(REDIS_CONTAINER) || docker run --name $(REDIS_CONTAINER) -d -p 6379:6379 redis:alpine
	@echo "Redis is running"

# Stop Redis container
redis-stop:
	@echo "Stopping Redis container..."
	@docker stop $(REDIS_CONTAINER) || true
	@docker rm $(REDIS_CONTAINER) || true
	@echo "Redis container stopped"

# Run the control plane with Redis
run: redis-start build
	./$(BINARY_NAME)

# Clean up
clean: redis-stop
	go clean
	rm -f $(BINARY_NAME)

# Run tests
test: redis-start
	go test ./...


# Build for production
prod:
	go build -o $(BINARY_NAME) -ldflags="-s -w" .

# Install dependencies
deps:
	go mod download

# Update dependencies
update:
	go get -u ./...
	go mod tidy 