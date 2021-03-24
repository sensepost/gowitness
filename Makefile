# ref: https://vic.demuzere.be/articles/golang-makefile-crosscompile/

G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
APPVER := $(shell grep 'version =' cmd/version.go | cut -d \" -f2)
PWD := $(shell pwd)
LD_FLAGS := -ldflags="-s -w -X=github.com/sensepost/gowitness/cmd.gitHash=$(V) -X=github.com/sensepost/gowitness/cmd.goVer=$(G)"
BIN_DIR := build
DOCKER_GO_VER := 1.15.10# https://github.com/elastic/golang-crossbuild
DOCKER_RELEASE_BUILD_CMD := docker run --rm -it -v $(PWD):/go/src/github.com/sensepost/gowitness \
	-w /go/src/github.com/sensepost/gowitness -e CGO_ENABLED=1 \
	docker.elastic.co/beats-dev/golang-crossbuild:$(DOCKER_GO_VER)

export CGO_ENABLED=1

default: clean generate darwin linux windows integrity

clean:
	$(RM) $(BIN_DIR)/gowitness*
	go clean -x

install:
	go install

generate:
	cd web && go generate && cd -

darwin:
	GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-darwin-amd64'
linux:
	GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-linux-amd64'
windows:
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-windows-amd64.exe'

# release
release: clean generate darwin-release linux-release windows-release integrity

darwin-release:
	$(DOCKER_RELEASE_BUILD_CMD)-darwin --build-cmd "make darwin" -p "darwin/amd64"
linux-release:
	$(DOCKER_RELEASE_BUILD_CMD)-main --build-cmd "make linux" -p "linux/amd64"
windows-release:
	$(DOCKER_RELEASE_BUILD_CMD)-main --build-cmd "make windows" -p "windows/amd64"

docker:
	go build $(LD_FLAGS) -o gowitness
docker-image:
	docker build -t gowitness:local .

integrity:
	cd $(BIN_DIR) && shasum *
