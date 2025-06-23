.PHONY: build run test test-cover test-cover-html test-race benchmark clean docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=go-url-shortener
DOCKER_IMAGE=go-url-shortener

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

run-memory:
	./$(BINARY_NAME) --db memory

run-sqlite:
	./$(BINARY_NAME) --db sqlite --db-path urls.db

test:
	$(GOTEST) -v ./...

test-cover:
	$(GOTEST) -v -cover ./...

test-cover-html:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-race:
	$(GOTEST) -v -race ./...

benchmark:
	$(GOTEST) -bench=. ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f urls.db

tidy:
	$(GOMOD) tidy

docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

docker-run:
	docker run -p 8080:8080 $(DOCKER_IMAGE):latest
