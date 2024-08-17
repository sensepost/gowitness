# ref: https://vic.demuzere.be/articles/golang-makefile-crosscompile/

G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
APPVER := $(shell grep 'version =' cmd/version.go | cut -d \" -f2)
PWD := $(shell pwd)
LD_FLAGS := -ldflags="-s -w -X=github.com/sensepost/gowitness/cmd.gitHash=$(V) -X=github.com/sensepost/gowitness/cmd.goVer=$(G)"
BIN_DIR := build
DOCKER_GO_VER := 1.22.6# https://github.com/elastic/golang-crossbuild
DOCKER_RELEASE_BUILD_CMD := docker run --rm -it -v $(PWD):/go/src/github.com/sensepost/gowitness \
	-w /go/src/github.com/sensepost/gowitness -e CGO_ENABLED=1 \
	docker.elastic.co/beats-dev/golang-crossbuild:$(DOCKER_GO_VER)

export CGO_ENABLED=1

default: clean darwin linux windows integrity

clean:
	$(RM) $(BIN_DIR)/gowitness*
	go clean -x

install:
	go install

darwin:
	GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-darwin-amd64'
darwin-arm:
	GOOS=darwin GOARCH=arm64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-darwin-arm64'
linux:
	GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-linux-amd64'
linux-arm:
	GOOS=linux GOARCH=arm64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-linux-arm64'
linux-armhf:
	GOOS=linux GOARCH=arm GOARM=7 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-linux-armv7'
windows:
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-windows-amd64.exe'

# release
release: clean darwin-release linux-release windows-release integrity

darwin-release:
	$(DOCKER_RELEASE_BUILD_CMD)-darwin-debian10 --build-cmd "make darwin" -p "darwin/amd64"
	$(DOCKER_RELEASE_BUILD_CMD)-darwin-arm64-debian10 --build-cmd "make darwin-arm" -p "darwin/arm64"
linux-release:
	$(DOCKER_RELEASE_BUILD_CMD)-main --build-cmd "make linux" -p "linux/amd64"
	$(DOCKER_RELEASE_BUILD_CMD)-arm --build-cmd "make linux-arm" -p "linux/arm64"
	$(DOCKER_RELEASE_BUILD_CMD)-armhf --build-cmd "make linux-armhf" -p "linux/armv7"
windows-release:
	$(DOCKER_RELEASE_BUILD_CMD)-main --build-cmd "make windows" -p "windows/amd64"

docker:
	go build $(LD_FLAGS) -o gowitness
docker-image:
	docker build -t gowitness:local .

integrity:
	cd $(BIN_DIR) && shasum *
