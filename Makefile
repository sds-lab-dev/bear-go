.PHONY: build run test fmt fmt-check staticcheck clean

BUILD_VERSION_SCRIPT := ./tools/build/build_version.sh

build:
	export BUILD_VERSION="$$(bash $(BUILD_VERSION_SCRIPT))"; \
    go build $(BUILD_FLAGS) -x -v -ldflags "-X main.buildVersion=$$BUILD_VERSION"

run:
	go run $(RUN_FLAGS) $(BUILD_LDFLAGS) ./...

test:
	go test $(TEST_FLAGS) ./...

fmt:
	go fmt $(FMT_FLAGS) ./...

fmt-check:
	test -z "$$(gofmt -l .)"

staticcheck:
	staticcheck $(STATICCHECK_FLAGS) ./...

clean:
	go clean -cache -modcache $(CLEAN_FLAGS)