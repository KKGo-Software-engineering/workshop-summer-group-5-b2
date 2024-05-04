PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...


PHONY: upload
upload:
	@echo "Uploading images..."
	curl -X POST http://localhost:8080/api/v1/upload \
	-H "Content-Type: multipart/form-data" \
	-F "images=@e-slip1.png" \
	-F "images=@e-slip2.png"

.PHONY: run
run:
	@echo "Running the server..."
	go run main.go

.PHONY: slow
slow:
	@echo "Running the server with slow response..."
	curl http://localhost:8080/api/v1/slow

.PHONY: health
health:
	@echo "Checking the health of the server..."
	curl http://localhost:8080/api/v1/health

.PHONY: users
users:
	@echo "Getting the users..."
	curl http://localhost:8080/api/v1/users