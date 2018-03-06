SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GO ?= go

# Install all the build and lint dependencies
setup:
	@$(GO) get -u github.com/alecthomas/gometalinter
	@$(GO) get -u github.com/golang/dep/cmd/dep
	@$(GO) get -u github.com/pierrre/gotestcover
	@$(GO) get -u golang.org/x/tools/cmd/cover
	@gometalinter --install
	@dep ensure
.PHONY: setup

# Install from source.
install:
	@echo "==> Installing up ${GOPATH}/bin/syslog-cloudlogs"
	@$(GO) install ./...
.PHONY: install

# Run all the tests
test:
	@gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	@$(GO) tool cover -html=coverage.txt
.PHONY: cover

# Run all the linters
lint:
	@gometalinter --vendor ./...
.PHONY: lint

# Run all the tests and code checks
ci: setup test lint
.PHONY: ci

# Release binaries to GitHub.
release:
	@echo "==> Releasing"
	@goreleaser -p 1 --rm-dist -config .goreleaser.yml
	@echo "==> Complete"
.PHONY: release

# Release binaries to GitHub.
snapshot:
	@echo "==> Releasing snapshot"
	@goreleaser --snapshot --rm-dist --debug -config .goreleaser.yml
	@echo "==> Complete"
.PHONY: snapshot
