.PHONY: build test lint clean

TMP_CACHE_ROOT ?= /tmp/doug-cache
GOCACHE_DIR := $(TMP_CACHE_ROOT)/go-build
GOLANGCI_LINT_CACHE_DIR := $(TMP_CACHE_ROOT)/golangci-lint
GOFMT_CHECK = files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		printf '%s\n' "$$files"; \
		exit 1; \
	fi

build:
	go build -o doug-plan .

test:
	go test ./...

lint:
	@mkdir -p "$(GOCACHE_DIR)" "$(GOLANGCI_LINT_CACHE_DIR)"
	@$(GOFMT_CHECK)
	GOCACHE="$(GOCACHE_DIR)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE_DIR)" golangci-lint run ./...
	GOCACHE="$(GOCACHE_DIR)" go vet ./...

clean:
	rm -f doug-plan
