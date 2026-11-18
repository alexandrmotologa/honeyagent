.PHONY: all build build-web test clean seed simulate

BINARY_NAME=honeyagent
BIN_DIR=bin

all: build-web build

build-web:
	@echo "Building React 19 Studio..."
	cd web && npm install && npm run build
	@echo "Embedding web assets into Go static folder..."
	mkdir -p internal/server/static
	cp -r web/dist/* internal/server/static/

build:
	@echo "Compiling zero-CGO Go binary..."
	mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/honeyagent

test:
	@echo "Running unit and integration test suite..."
	go test -v -race ./...

seed:
	@echo "Seeding default memory vector store..."
	./$(BIN_DIR)/$(BINARY_NAME) seed --store memory

simulate:
	@echo "Running adversarial attack simulation..."
	./$(BIN_DIR)/$(BINARY_NAME) simulate --type canary_exfiltration

clean:
	rm -rf $(BIN_DIR) web/dist
