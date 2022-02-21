.PHONY: clean build build-mac build-linux

export GO111MODULE=on
export CGO_ENABLED=0

PACKAGE_FOLDER = gitlab-code-quality

# Build Flags
VERSION ?= $(VERSION:)
BUILD_NUMBER ?= $(BUILD_NUMBER:)
BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)

# If we don't set the build number it defaults to dev
ifeq ($(VERSION),)
	VERSION := 0.0.0
endif

# If we don't set the build number it defaults to dev
ifeq ($(BUILD_NUMBER),)
	BUILD_NUMBER := dev
endif

LDFLAGS += -X "$(PACKAGE_FOLDER)/model.Version=$(VERSION)"
LDFLAGS += -X "$(PACKAGE_FOLDER)/model.BuildNumber=$(BUILD_NUMBER)"
LDFLAGS += -X "$(PACKAGE_FOLDER)/model.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "$(PACKAGE_FOLDER)/model.BuildHash=$(BUILD_HASH)"

all: build

build:
	go build -ldflags '$(LDFLAGS)' -o ./bin/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

dep-install:
	go mod download

build-mac:
	mkdir -p bin/mac
	$(eval LDFLAGS += -X "$(PACKAGE_FOLDER)/model.Edition=mac")
	env GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o ./bin/mac/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

build-linux:
	mkdir -p bin/linux
	$(eval LDFLAGS += -X "$(PACKAGE_FOLDER)/model.Edition=linux")
	env GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o ./bin/linux/$(PACKAGE_FOLDER) ./cmd/$(PACKAGE_FOLDER)/main.go

package:
	@ echo Packaging
	@# Remove any old files
	rm -Rf $(DIST_ROOT)

	@# Create needed directories
	mkdir -p $(DIST_PATH)/bin

	@# Package webapp
	mkdir -p $(DIST_PATH)/client
	cp -RL $(BUILD_WEBAPP_DIR)/${DIST_ROOT}/* $(DIST_PATH)/client

linux-package: package server-linux

lint: check-lint
	golangci-lint run ./...

lint-diff: check-lint
	golangci-lint run ./... --out-format tab --new-from-rev=HEAD~1

lint-checkstyle: check-lint
	golangci-lint --out-format checkstyle run ./...

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