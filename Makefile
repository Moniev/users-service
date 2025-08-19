GO=go

# ==============================================================================
# Testing Targets
# ==============================================================================
test: test-unit test-integration test-e2e

test-unit:
	@echo "--- Running Standard Unit Tests (in app/) ---"
	@$(GO) test -v -tags=unit ./app/...
	@echo "\n--- Running Data-Driven Unit Tests (in tests/) ---"
	@$(GO) test -v -tags=unit ./tests -args -test_type=unit
	@swag init -g ./app/cmd/main.go -o app/docs

test-integration:
	@echo "--- Running Integration Tests (requires Docker) ---"
	@$(GO) test -v -tags=integration ./app/...
	@$(GO) test -v -tags=integration ./tests -args -test_type=integration
	@swag init -g ./app/cmd/main.go -o app/docs
    @docker build -t ghcr.io/factory-chainline/users-service:latest -f ./docker/Dockerfile .

test-e2e:
	@echo "--- Running End-to-End Tests (requires Docker) ---"
	@$(GO) test -v -tags=e2e ./app/...
	@$(GO) test -v -tags=e2e ./tests -args -test_type=e2e
	@swag init -g ./app/cmd/main.go -o app/docs
	@docker build -t ghcr.io/factory-chainline/users-service:latest -f ./docker/Dockerfile .


# ==============================================================================
# Code Generation Targets
# ==============================================================================
generate-ent:
	@echo "--- Generating Ent schema models ---"
	@$(GO) run entgo.io/ent/cmd/ent generate ./app/models/ent/schema
	@swag init -g ./app/cmd/main.go -o app/docs

generate-mocks:
	@echo "--- Generating mocks with mockery ---"
	@mockery
	@swag init -g app/cmd/main.go -o app/docs

generate-swagger:
	@echo "--- Generating Swagger documentation ---"
	@swag init -g ./app/cmd/main.go -o ./app/docs


# ==============================================================================
# Build Targets
# ==============================================================================
build-docker:
	@echo "--- Building Docker image ---"
	@docker build -t ghcr.io/factory-chainline/users-service:latest -f ./docker/Dockerfile .

.PHONY: test test-unit test-integration test-e2e generate-ent generate-mocks generate-swagger build-docker
