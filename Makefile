.PHONY: clean build build-mac build-linux

PACKAGE_FOLDER = gitlab-reporter

export GO111MODULE=on
export CGO_ENABLED=0

# Strip debug info
LDFLAGS += "-s -w"

all: build

build:
	go build -ldflags $(LDFLAGS) -o ./bin/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

dep-install:
	go mod download

build-mac:
	mkdir -p bin/mac
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o ./bin/mac/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

build-linux:
	mkdir -p bin/linux
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o ./bin/linux/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

lint: check-lint
	golangci-lint run ./...

lint-diff: check-lint
	golangci-lint run ./... --out-format tab --new-from-rev=HEAD~1

lint-checkstyle: check-lint
	golangci-lint --out-format checkstyle run ./...

start:
	go run cmd/gitlab-reporter/main.go transform

test:
	go test -v ./...

doc:
	go doc ./...

check-lint:
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "Downloading golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi;

docker:
	docker build -f Dockerfile -t ${PACKAGE_FOLDER} .

clean:
	rm -rf bin
