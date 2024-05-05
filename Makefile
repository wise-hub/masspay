APP_NAME := masspay
BINARY_DIR := bin
DOCKER_IMAGE := masspay:latest
DOCKER_IMAGE_TAR := masspay.tar
DOCKER_COMPOSE_FILE := docker-compose.yml

clean:
	@echo "Cleaning up binaries..."
	rm -rf $(BINARY_DIR)

build_app: clean
	@echo "Building Go application..."
	mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_DIR)/$(APP_NAME) ./cmd/main.go

build_image: build_app
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

export_image: build_image
	@echo "Exporting Docker image to tarball..."
	docker save -o $(DOCKER_IMAGE_TAR) $(DOCKER_IMAGE)

clean_containers:
	@echo "Stopping and removing existing containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

build_containers: clean_containers build_image
	@echo "Building containers using Docker Compose..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d

run_local: build_app
	@echo "Running application locally..."
	./$(BINARY_DIR)/$(APP_NAME)

all: build_app build_image build_containers

.PHONY: clean build_app build_image build_containers clean_containers export_image run_local all
