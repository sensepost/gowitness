G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
APPVER := $(shell grep 'Version =' internal/version/version.go | cut -d \" -f2)
PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LD_FLAGS := -trimpath \
	-ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.GitHash=$(V) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildEnv=$(G) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildTime=$(BUILD_TIME)"
BIN_DIR := build
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64
CGO := CGO_ENABLED=0

# Default target
default: clean test frontend api-doc build integrity

# Clean up build artifacts
clean:
	find $(BIN_DIR) -type f -name 'gowitness-*' -delete || true
	go clean -x

# Build frontend
frontend: check-npm
	@echo "Building frontend..."
	cd web/ui && npm i && npm run build

# Check if npm is installed
check-npm:
	@command -v npm >/dev/null 2>&1 || { echo >&2 "npm is not installed. Please install npm first."; exit 1; }

# Generate a swagger.json used for the api documentation
api-doc:
	go install github.com/swaggo/swag/cmd/swag@latest
	$(GOPATH)/bin/swag i --exclude ./web/ui --output web/docs
	$(GOPATH)/bin/swag f

# Run any tests
test:
	@echo "Running tests..."
	go test ./...

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
