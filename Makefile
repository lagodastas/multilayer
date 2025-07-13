#make build    # Собрать приложение
#make run      # Запустить приложение - бинарник
#run-dev	   # Запустить приложение - go run ./cmd/server/main.
#make test     # Запустить тесты
#make lint     # Проверить код линтером
#make clean    # Очистить артефакты сборки
#make help     # Показать справку
#make docker-build # Собрать Docker образ
#make docker-run   # Запустить в Docker
#make docker-compose-dev # Запустить в Docker Compose (dev)
#make docker-compose-prod # Запустить в Docker Compose (prod)
#make k8s-deploy # Развернуть в Kubernetes
#make k8s-deploy-local # Развернуть в локальный Kubernetes
#make k8s-undeploy # Удалить из Kubernetes

# Project variables
BINARY_NAME := multilayer
#GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_FILES := $(shell powershell -Command "Get-ChildItem -Recurse -Filter '*.go' -Exclude 'vendor' | ForEach-Object { $_.FullName }")
DOCKER_IMAGE := multilayer-app

.PHONY: all build clean test lint run help docker-build docker-run docker-clean docker-compose-dev docker-compose-prod k8s-deploy k8s-deploy-local k8s-undeploy

all: build

## Build the application
build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME).exe ./cmd/server
#	@go build -o $(BINARY_NAME) ./cmd/server

## Run the application
run: build
	@echo "Starting application..."
	@if exist $(BINARY_NAME).exe ./$(BINARY_NAME).exe
	@if exist $(BINARY_NAME) ./$(BINARY_NAME)


run-dev:
	@echo "Starting application with go run..."
	go run ./cmd/server/main.go

## Run tests
test:
	@echo "Running tests..."
	@go test -v -cover  -count=1 -race  ./...

## Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -count=1 -race -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@golangci-lint run

## Clean build artifacts
clean:
	@echo "Cleaning up..."
	@go clean
	@if exist $(BINARY_NAME).exe del $(BINARY_NAME).exe
	@if exist $(BINARY_NAME) rm $(BINARY_NAME)
	@if exist coverage.out rm coverage.out
	@if exist coverage.html del coverage.html

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

## Build Docker image (production)
docker-build-prod:
	@echo "Building production Docker image..."
	@docker build -f Dockerfile.prod -t $(DOCKER_IMAGE)-prod .

## Run in Docker
docker-run: docker-build
	@echo "Running in Docker..."
	@docker run -p 8080:8080 --name $(DOCKER_IMAGE)-container $(DOCKER_IMAGE)

## Clean Docker artifacts
docker-clean:
	@echo "Cleaning Docker artifacts..."
	@docker stop $(DOCKER_IMAGE)-container 2>nul || echo "Container not running"
	@docker rm $(DOCKER_IMAGE)-container 2>nul || echo "Container not found"
	@docker rmi $(DOCKER_IMAGE) 2>nul || echo "Image not found"

## Run with Docker Compose (development)
docker-compose-dev:
	@echo "Starting development environment..."
	@docker-compose -f docker-compose.dev.yml up --build

## Run with Docker Compose (production)
docker-compose-prod:
	@echo "Starting production environment..."
	@docker-compose up --build

## Stop Docker Compose
docker-compose-down:
	@echo "Stopping Docker Compose..."
	@docker-compose down
	@docker-compose -f docker-compose.dev.yml down

## Deploy to Kubernetes
k8s-deploy: docker-build-prod
	@echo "Deploying to Kubernetes..."
	@cd k8s && ./deploy.sh

## Deploy to local Kubernetes cluster
k8s-deploy-local: docker-build-prod
	@echo "Deploying to local Kubernetes cluster..."
	@cd k8s && ./deploy-local.sh

## Undeploy from Kubernetes
k8s-undeploy:
	@echo "Undeploying from Kubernetes..."
	@cd k8s && ./undeploy.sh

## Show Kubernetes status
k8s-status:
	@echo "Kubernetes deployment status:"
	@kubectl get pods -n multilayer
	@echo ""
	@echo "Services:"
	@kubectl get svc -n multilayer
	@echo ""
	@echo "Ingress:"
	@kubectl get ingress -n multilayer

## Show help
help:
	@echo "Available targets:"
	@echo "  build              - Compile the application"
	@echo "  run                - Build and run the application"
	@echo "  test               - Run tests"
	@echo "  test-cover         - Run tests with coverage report"
	@echo "  lint               - Run linters"
	@echo "  clean              - Remove build artifacts"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Build and run in Docker"
	@echo "  docker-clean       - Clean Docker artifacts"
	@echo "  docker-compose-dev - Run with Docker Compose (dev)"
	@echo "  docker-compose-prod- Run with Docker Compose (prod)"
	@echo "  docker-compose-down- Stop Docker Compose"
	@echo "  k8s-deploy         - Deploy to Kubernetes"
	@echo "  k8s-deploy-local   - Deploy to local Kubernetes cluster"
	@echo "  k8s-undeploy       - Undeploy from Kubernetes"
	@echo "  k8s-status         - Show Kubernetes status"
	@echo "  help               - Show this help message"


## Run migrations
#migrate:
#	@go run cmd/migrate/main.go
