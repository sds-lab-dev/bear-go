.PHONY: build run test staticcheck clean

BUILD_VERSION_SCRIPT := ./.github/scripts/build_version.sh
BUILD_VERSION := $(shell $(BUILD_VERSION_SCRIPT))
BUILD_LDFLAGS := -ldflags '-X main.buildVersion=$(BUILD_VERSION)'

build:
	go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) ./...

run:
	go run $(RUN_FLAGS) $(BUILD_LDFLAGS) ./...

test:
	go test $(TEST_FLAGS) -v ./...

staticcheck:
	staticcheck $(STATICCHECK_FLAGS) ./...

clean:
	go clean $(CLEAN_FLAGS) ./...