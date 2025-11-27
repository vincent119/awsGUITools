GOFMT=goimports
APP=aws-tui

.PHONY: all lint test build run tidy

all: lint test build

lint:
	@echo "==> running gofmt/goimports/go vet"
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')
	goimports -w $(shell find . -name '*.go' -not -path './vendor/*')
	go vet ./...

test:
	@echo "==> running tests"
	go test ./...

build:
	@echo "==> building $(APP)"
	go build -o bin/$(APP) ./cmd/aws-tui

run: build
	./bin/$(APP)

tidy:
	go mod tidy

