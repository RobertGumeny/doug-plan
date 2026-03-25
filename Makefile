.PHONY: build build-ui fmt fmt-check test test-integration lint clean

TMP_CACHE_ROOT ?= /tmp/doug-cache
GOCACHE_DIR := $(TMP_CACHE_ROOT)/go-build
GOLANGCI_LINT_CACHE_DIR := $(TMP_CACHE_ROOT)/golangci-lint
GOFMT_CHECK = files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		printf '%s\n' "$$files"; \
		exit 1; \
	fi

build-ui:
	npm install --prefix ui
	node ui/build.js

build: build-ui
	go build -o doug-plan .

fmt:
	gofmt -w .

fmt-check:
	@$(GOFMT_CHECK)

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	@mkdir -p "$(GOCACHE_DIR)" "$(GOLANGCI_LINT_CACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE_DIR)" golangci-lint run ./...
	GOCACHE="$(GOCACHE_DIR)" go vet ./...

clean:
	rm -f doug-plan
