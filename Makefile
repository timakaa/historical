# Default variable
SERVICE ?= all

# Run services
run:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service..." && \
			cd prices && go run cmd/main.go \
			;; \
		"auth") \
			echo "Starting auth service..." && \
			cd auth && go run cmd/main.go \
			;; \
		"gateway") \
			echo "Starting Gateway service..." && \
			cd gateway && go run cmd/main.go \
			;; \
		"all") \
			echo "Starting all services..." && \
			(cd gateway && go run cmd/main.go) \
			(cd prices && go run cmd/main.go) & \
			(cd auth && go run cmd/main.go) & \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run all services (alternative method)
run-all:
	@echo "Starting all services..."
	(cd gateway && go run cmd/main.go) & \
	@(cd prices && go run cmd/main.go) & \
	(cd auth && go run cmd/main.go)

# Build services
build:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Building Prices service..." && \
			cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go \
			;; \
		"auth") \
			echo "Building auth service..." && \
			cd auth && mkdir -p bin && go build -o bin/auth cmd/main.go \
			;; \
		"gateway") \
			echo "Building Gateway service..." && \
			cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go \
			;; \
		"all") \
			echo "Building all services..." && \
			(cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go) && \
			(cd auth && mkdir -p bin && go build -o bin/auth cmd/main.go) && \
			(cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Build all services (alternative method)
build-all:
	@echo "Building all services..."
	@cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go
	@cd auth && mkdir -p bin && go build -o bin/auth cmd/main.go
	@cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go
	@echo "All services built in their respective bin directories"

# Run built binaries
start:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service..." && \
			cd prices && ./bin/prices \
			;; \
		"auth") \
			echo "Starting auth service..." && \
			cd auth && ./bin/auth \
			;; \
		"gateway") \
			echo "Starting Gateway service..." && \
			cd gateway && ./bin/gateway \
			;; \
		"all") \
			echo "Starting all services..." && \
			(cd prices && ./bin/prices) & \
			(cd auth && ./bin/auth) & \
			(cd gateway && ./bin/gateway) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run all built binaries (alternative method)
start-all:
	@echo "Starting all services from binaries..."
	@(cd prices && ./bin/prices) & \
	(cd auth && ./bin/auth) & \
	(cd gateway && ./bin/gateway)

# Generate proto files
gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(shell find common/proto -name "*.proto")

# Stop all services
stop:
	@echo "Stopping all services..."
	@-pkill -f "go run cmd/main.go" 2>/dev/null || true
	@-pkill -f "bin/prices" 2>/dev/null || true
	@-pkill -f "bin/auth" 2>/dev/null || true
	@-pkill -f "bin/gateway" 2>/dev/null || true
	@-pkill -f "air" 2>/dev/null || true
	@-lsof -ti :50050,50051,50052 | xargs kill -9 2>/dev/null || true
	@echo "All services stopped."

# Clean binaries
clean:
	@echo "Cleaning binaries..."
	@rm -rf prices/bin auth/bin gateway/bin
	@echo "Binaries cleaned."

# Help
help:
	@echo "Available commands:"
	@echo "  make run SERVICE=prices      - Run Prices service"
	@echo "  make run SERVICE=auth      - Run auth service"
	@echo "  make run SERVICE=gateway     - Run Gateway service"
	@echo "  make run SERVICE=all         - Run all services"
	@echo "  make run-all                 - Run all services (alternative)"
	@echo "  make build SERVICE=prices    - Build Prices service"
	@echo "  make build SERVICE=auth    - Build auth service"
	@echo "  make build SERVICE=gateway   - Build Gateway service"
	@echo "  make build SERVICE=all       - Build all services"
	@echo "  make build-all               - Build all services (alternative)"
	@echo "  make start SERVICE=prices    - Start Prices binary"
	@echo "  make start SERVICE=auth    - Start auth binary"
	@echo "  make start SERVICE=gateway   - Start Gateway binary"
	@echo "  make start SERVICE=all       - Start all binaries"
	@echo "  make start-all               - Start all binaries (alternative)"
	@echo "  make gen                     - Generate proto files"
	@echo "  make stop                    - Stop all services"
	@echo "  make clean                   - Clean binaries"
	@echo "  make test SERVICE=prices     - Run all tests for Prices service"
	@echo "  make test SERVICE=auth       - Run all tests for Auth service"
	@echo "  make test SERVICE=gateway    - Run all tests for Gateway service"
	@echo "  make test SERVICE=all        - Run all tests for all services"
	@echo "  make test-unit SERVICE=prices - Run unit tests for Prices service"
	@echo "  make test-unit SERVICE=auth  - Run unit tests for Auth service"
	@echo "  make test-unit SERVICE=gateway - Run unit tests for Gateway service"
	@echo "  make test-unit SERVICE=all   - Run unit tests for all services"
	@echo "  make test-integration SERVICE=prices - Run integration tests for Prices service"
	@echo "  make test-integration SERVICE=auth - Run integration tests for Auth service"
	@echo "  make test-integration SERVICE=gateway - Run integration tests for Gateway service"
	@echo "  make test-integration SERVICE=all - Run integration tests for all services"
	@echo "  make test-coverage SERVICE=prices - Run tests with coverage for Prices service"
	@echo "  make test-coverage SERVICE=auth - Run tests with coverage for Auth service"
	@echo "  make test-coverage SERVICE=gateway - Run tests with coverage for Gateway service"
	@echo "  make test-coverage SERVICE=all - Run tests with coverage for all services"
	@echo "  make test-unit-coverage SERVICE=prices - Run unit tests with coverage for Prices service"
	@echo "  make test-unit-coverage SERVICE=auth - Run unit tests with coverage for Auth service"
	@echo "  make test-unit-coverage SERVICE=gateway - Run unit tests with coverage for Gateway service"
	@echo "  make test-unit-coverage SERVICE=all - Run unit tests with coverage for all services"
	@echo "  make test-integration-coverage SERVICE=prices - Run integration tests with coverage for Prices service"
	@echo "  make test-integration-coverage SERVICE=auth - Run integration tests with coverage for Auth service"
	@echo "  make test-integration-coverage SERVICE=gateway - Run integration tests with coverage for Gateway service"
	@echo "  make test-integration-coverage SERVICE=all - Run integration tests with coverage for all services"
	@echo "  make help                    - Show this help"

# Run services with Air for automatic reload
dev:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service with Air..." && \
			cd prices && air \
			;; \
		"auth") \
			echo "Starting auth service with Air..." && \
			cd auth && air \
			;; \
		"gateway") \
			echo "Starting Gateway service with Air..." && \
			cd gateway && air \
			;; \
		"all") \
			echo "Starting all services with Air..." && \
			( trap 'kill 0' SIGINT SIGTERM EXIT; \
			  (cd prices && air) & \
			  (cd auth && air) & \
			  (cd gateway && air) & \
			  wait \
			) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run all services with Air (alternative method)
dev-all:
	@echo "Starting all services with Air..."
	@( trap 'kill 0' SIGINT SIGTERM EXIT; \
	   (cd prices && air) & \
	   (cd auth && air) & \
	   (cd gateway && air) & \
	   wait \
	)

# Run tests
test:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running all tests for Prices service..." && \
			cd prices && go test -v ./internal/... \
			;; \
		"auth") \
			echo "Running all tests for Auth service..." && \
			cd auth && go test -v ./internal/... \
			;; \
		"gateway") \
			echo "Running all tests for Gateway service..." && \
			cd gateway && go test -v ./internal/... \
			;; \
		"all") \
			echo "Running all tests for all services..." && \
			(cd prices && go test -v ./internal/...) && \
			(cd auth && go test -v ./internal/...) && \
			(cd gateway && go test -v ./internal/...) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run only unit tests
test-unit:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running unit tests for Prices service..." && \
			cd prices && go test -v ./internal/... -run "Unit$$" \
			;; \
		"auth") \
			echo "Running unit tests for Auth service..." && \
			cd auth && go test -v ./internal/... -run "Unit$$" \
			;; \
		"gateway") \
			echo "Running unit tests for Gateway service..." && \
			cd gateway && go test -v ./internal/... -run "Unit$$" \
			;; \
		"all") \
			echo "Running unit tests for all services..." && \
			(cd prices && go test -v ./internal/... -run "Unit$$") && \
			(cd auth && go test -v ./internal/... -run "Unit$$") && \
			(cd gateway && go test -v ./internal/... -run "Unit$$") \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run only integration tests
test-integration:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running integration tests for Prices service..." && \
			cd prices && go test -v ./internal/... -run "^Test[^Unit]" \
			;; \
		"auth") \
			echo "Running integration tests for Auth service..." && \
			cd auth && go test -v ./internal/... -run "^Test[^Unit]" \
			;; \
		"gateway") \
			echo "Running integration tests for Gateway service..." && \
			cd gateway && go test -v ./internal/... -run "^Test[^Unit]" \
			;; \
		"all") \
			echo "Running integration tests for all services..." && \
			(cd prices && go test -v ./internal/... -run "^Test[^Unit]") && \
			(cd auth && go test -v ./internal/... -run "^Test[^Unit]") && \
			(cd gateway && go test -v ./internal/... -run "^Test[^Unit]") \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run tests with code coverage
test-coverage:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running tests with coverage for Prices service..." && \
			cd prices && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"auth") \
			echo "Running tests with coverage for Auth service..." && \
			cd auth && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"gateway") \
			echo "Running tests with coverage for Gateway service..." && \
			cd gateway && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"all") \
			echo "Running tests with coverage for all services..." && \
			(cd prices && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd auth && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd gateway && go test -v ./internal/... -coverprofile=coverage.out -coverpkg=./internal/...) && \
			echo "Coverage reports generated in each service directory" \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run only unit tests with code coverage
test-unit-coverage:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running unit tests with coverage for Prices service..." && \
			cd prices && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"auth") \
			echo "Running unit tests with coverage for Auth service..." && \
			cd auth && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"gateway") \
			echo "Running unit tests with coverage for Gateway service..." && \
			cd gateway && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"all") \
			echo "Running unit tests with coverage for all services..." && \
			(cd prices && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd auth && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd gateway && go test -v ./internal/... -run "Unit$$" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			echo "Unit test coverage reports generated in each service directory" \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

# Run only integration tests with code coverage
test-integration-coverage:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Running integration tests with coverage for Prices service..." && \
			cd prices && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"auth") \
			echo "Running integration tests with coverage for Auth service..." && \
			cd auth && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"gateway") \
			echo "Running integration tests with coverage for Gateway service..." && \
			cd gateway && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/... && go tool cover -html=coverage.out \
			;; \
		"all") \
			echo "Running integration tests with coverage for all services..." && \
			(cd prices && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd auth && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			(cd gateway && go test -v ./internal/... -run "^Test[^Unit]" -coverprofile=coverage.out -coverpkg=./internal/...) && \
			echo "Integration test coverage reports generated in each service directory" \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, auth, gateway, all" && \
			exit 1 \
			;; \
	esac

.PHONY: run run-all build build-all start start-all gen stop clean help dev dev-all test test-unit test-integration test-coverage test-unit-coverage test-integration-coverage