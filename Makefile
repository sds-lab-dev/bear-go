.PHONY: build run test fmt fmt-check staticcheck clean

BUILD_VERSION_SCRIPT := ./tools/build/build_version.sh

build:
	export BUILD_VERSION="$$(bash $(BUILD_VERSION_SCRIPT))"; \
    go build $(BUILD_FLAGS) -v -ldflags "-X main.buildVersion=$$BUILD_VERSION"

run:
	go run $(RUN_FLAGS) $(BUILD_LDFLAGS) ./...

test:
	go test $(TEST_FLAGS) ./...

fmt:
	go fmt $(FMT_FLAGS) ./...

fmt-check:
	@echo "go version: $$(go version)"
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "Unformatted Go files:"; \
		echo "$$files"; \
		exit 1; \
	fi

staticcheck:
	staticcheck $(STATICCHECK_FLAGS) ./...

clean:
	go clean -cache -modcache $(CLEAN_FLAGS)