.PHONY: docker-up
docker-up:
	@echo "Starting Docker Compose..."
	docker-compose -f docker-compose.yaml up -d

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker Compose..."
	docker-compose -f docker-compose.yaml down

init:
	@echo "Initializing dependencies..."
	cd API-service && go mod download
	cd sent-mail-service && go mod download
	cd reset-password-page && npm install

.PHONY: run-api-service
run-api-service:
	@echo "Running API Service..."
	cd API-service && go run ./main.go

.PHONY: run-mail-service
run-mail-service:
	@echo "Running Sent Mail Service..."
	cd sent-mail-service && go run ./main.go

.PHONY: run-fe
run-fe:
	@echo "Running Reset Password Page..."
	cd reset-password-page && npm start
