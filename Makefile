# Makefile for tcrs-cli

BINARY_NAME=tcrs
VERSION?=0.1.0
BUILD_DIR=build
INSTALL_DIR?=/usr/local/bin
SKILL_DIR?=$(HOME)/.claude/skills

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: all build clean install uninstall test skill cross-compile dev help

# Default target
all: build

# Build binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Install binary and skill
install: build skill
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

# Uninstall binary and skill
uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -rf $(SKILL_DIR)/tcrs
	@echo "Uninstalled $(BINARY_NAME)"

# Install Claude Code skill only
skill:
	mkdir -p $(SKILL_DIR)/tcrs
	cp skills/tcrs/SKILL.md $(SKILL_DIR)/tcrs/
	@echo "Installed skill to $(SKILL_DIR)/tcrs"

# Run tests
test:
	go test -v ./...

# Cross-compile for multiple platforms
cross-compile: clean
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Cross-compiled binaries in $(BUILD_DIR)/"

# Development build
dev: build
	@echo "Built $(BINARY_NAME) in current directory"

# Download dependencies
deps:
	go mod download
	go mod tidy

# Show help
help:
	@echo "TCRS CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build binary"
	@echo "  make install        - Install binary and skill"
	@echo "  make uninstall      - Remove installed files"
	@echo "  make skill          - Install Claude Code skill only"
	@echo "  make test           - Run tests"
	@echo "  make cross-compile  - Build for all platforms"
	@echo "  make clean          - Remove built files"
	@echo "  make dev            - Build for development"
	@echo "  make deps           - Download and tidy dependencies"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  INSTALL_DIR=$(INSTALL_DIR)"
	@echo "  SKILL_DIR=$(SKILL_DIR)"
