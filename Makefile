VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/massdriver-cloud/mcp-server/mcp.Version=$(VERSION)

.PHONY: test
test:
	go test ./... -cover -count=1

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/mcp-server .

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docker.build
docker.build:
	docker build --build-arg VERSION=$(VERSION) -t massdrivercloud/mcp-server .
