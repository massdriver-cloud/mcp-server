.PHONY: test
test:
	go test ./... -cover -count=1

.PHONY: build
build:
	go build -o bin/mcp-server .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docker.build
docker.build:
	docker build -t massdrivercloud/mcp-server .