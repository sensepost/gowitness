G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
APPVER := $(shell grep 'version =' internal/version/version.go | cut -d \" -f2)
PWD := $(shell pwd)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LD_FLAGS := -trimpath \
	-ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.gitHash=$(V) \
	-X=github.com/sensepost/gowitness/internal/version.goBuildEnv=$(G) \
	-X=github.com/sensepost/gowitness/internal/version.goBuildTime=$(BUILD_TIME)"
BIN_DIR := build
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm windows/amd64
CGO := CGO_ENABLED=0

# Default target
default: clean frontend build integrity

# Clean up build artifacts
clean:
	find $(BIN_DIR) -type f -name 'gowitness-*' -delete
	go clean -x

# Build frontend
frontend: check-npm
	@echo "Building frontend..."
	cd web/ui && npm i && npm run build

# Check if npm is installed
check-npm:
	@command -v npm >/dev/null 2>&1 || { echo >&2 "npm is not installed. Please install npm first."; exit 1; }


# Build for all platforms
build: $(PLATFORMS)

# Generic build target for platforms
$(PLATFORMS):
	$(eval GOOS=$(firstword $(subst /, ,$@)))
	$(eval GOARCH=$(lastword $(subst /, ,$@)))
	$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe)'

# Checksum integrity
integrity:
	cd $(BIN_DIR) && shasum *
