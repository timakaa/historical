# Переменная по умолчанию
SERVICE ?= all

# Запуск сервисов
run:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service..." && \
			cd prices && go run cmd/main.go \
			;; \
		"access") \
			echo "Starting Access service..." && \
			cd access && go run cmd/main.go \
			;; \
		"gateway") \
			echo "Starting Gateway service..." && \
			cd gateway && go run cmd/main.go \
			;; \
		"all") \
			echo "Starting all services..." && \
			(cd prices && go run cmd/main.go) & \
			(cd access && go run cmd/main.go) & \
			(cd gateway && go run cmd/main.go) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, access, gateway, all" && \
			exit 1 \
			;; \
	esac

# Запуск всех сервисов (альтернативный способ)
run-all:
	@echo "Starting all services..."
	@(cd prices && go run cmd/main.go) & \
	(cd access && go run cmd/main.go) & \
	(cd gateway && go run cmd/main.go)

# Сборка сервисов
build:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Building Prices service..." && \
			cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go \
			;; \
		"access") \
			echo "Building Access service..." && \
			cd access && mkdir -p bin && go build -o bin/access cmd/main.go \
			;; \
		"gateway") \
			echo "Building Gateway service..." && \
			cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go \
			;; \
		"all") \
			echo "Building all services..." && \
			(cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go) && \
			(cd access && mkdir -p bin && go build -o bin/access cmd/main.go) && \
			(cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, access, gateway, all" && \
			exit 1 \
			;; \
	esac

# Сборка всех сервисов (альтернативный способ)
build-all:
	@echo "Building all services..."
	@cd prices && mkdir -p bin && go build -o bin/prices cmd/main.go
	@cd access && mkdir -p bin && go build -o bin/access cmd/main.go
	@cd gateway && mkdir -p bin && go build -o bin/gateway cmd/main.go
	@echo "All services built in their respective bin directories"

# Запуск собранных бинарных файлов
start:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service..." && \
			cd prices && ./bin/prices \
			;; \
		"access") \
			echo "Starting Access service..." && \
			cd access && ./bin/access \
			;; \
		"gateway") \
			echo "Starting Gateway service..." && \
			cd gateway && ./bin/gateway \
			;; \
		"all") \
			echo "Starting all services..." && \
			(cd prices && ./bin/prices) & \
			(cd access && ./bin/access) & \
			(cd gateway && ./bin/gateway) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, access, gateway, all" && \
			exit 1 \
			;; \
	esac

# Запуск всех собранных бинарных файлов (альтернативный способ)
start-all:
	@echo "Starting all services from binaries..."
	@(cd prices && ./bin/prices) & \
	(cd access && ./bin/access) & \
	(cd gateway && ./bin/gateway)

# Генерация proto файлов
gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(shell find common/proto -name "*.proto")

# Остановка всех сервисов
stop:
	@echo "Stopping all services..."
	@-pkill -f "go run cmd/main.go" 2>/dev/null || true
	@-pkill -f "bin/prices" 2>/dev/null || true
	@-pkill -f "bin/access" 2>/dev/null || true
	@-pkill -f "bin/gateway" 2>/dev/null || true
	@echo "All services stopped."

# Очистка бинарных файлов
clean:
	@echo "Cleaning binaries..."
	@rm -rf prices/bin access/bin gateway/bin
	@echo "Binaries cleaned."

# Помощь
help:
	@echo "Available commands:"
	@echo "  make run SERVICE=prices      - Run Prices service"
	@echo "  make run SERVICE=access      - Run Access service"
	@echo "  make run SERVICE=gateway     - Run Gateway service"
	@echo "  make run SERVICE=all         - Run all services"
	@echo "  make run-all                 - Run all services (alternative)"
	@echo "  make build SERVICE=prices    - Build Prices service"
	@echo "  make build SERVICE=access    - Build Access service"
	@echo "  make build SERVICE=gateway   - Build Gateway service"
	@echo "  make build SERVICE=all       - Build all services"
	@echo "  make build-all               - Build all services (alternative)"
	@echo "  make start SERVICE=prices    - Start Prices binary"
	@echo "  make start SERVICE=access    - Start Access binary"
	@echo "  make start SERVICE=gateway   - Start Gateway binary"
	@echo "  make start SERVICE=all       - Start all binaries"
	@echo "  make start-all               - Start all binaries (alternative)"
	@echo "  make gen                     - Generate proto files"
	@echo "  make stop                    - Stop all services"
	@echo "  make clean                   - Clean binaries"
	@echo "  make help                    - Show this help"

# Запуск сервисов с Air для автоматической перезагрузки
dev:
	@case "$(SERVICE)" in \
		"prices") \
			echo "Starting Prices service with Air..." && \
			cd prices && air \
			;; \
		"access") \
			echo "Starting Access service with Air..." && \
			cd access && air \
			;; \
		"gateway") \
			echo "Starting Gateway service with Air..." && \
			cd gateway && air \
			;; \
		"all") \
			echo "Starting all services with Air..." && \
			(cd prices && air) & \
			(cd access && air) & \
			(cd gateway && air) \
			;; \
		*) \
			echo "Unknown service: $(SERVICE). Available options: prices, access, gateway, all" && \
			exit 1 \
			;; \
	esac

# Запуск всех сервисов с Air (альтернативный способ)
dev-all:
	@echo "Starting all services with Air..."
	@(cd prices && air) & \
	(cd access && air) & \
	(cd gateway && air)

.PHONY: run run-all build build-all start start-all gen stop clean help dev dev-all