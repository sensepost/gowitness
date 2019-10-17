# ref: https://vic.demuzere.be/articles/golang-makefile-crosscompile/

G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
LD_FLAGS := -ldflags="-s -w -X=github.com/sensepost/gowitness/cmd.gitHash=$(V) -X=github.com/sensepost/gowitness/cmd.goVer=$(G)"
BIN_DIR := build

default: clean darwin linux windows integrity

clean:
	$(RM) $(BIN_DIR)/gowitness*
	go clean -x

install:
	go install

darwin:
	GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-darwin-amd64'

linux:
	GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-linux-amd64'

windows:
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-windows-amd64.exe'

docker:
	go build $(LD_FLAGS) -o gowitness

integrity:
	cd $(BIN_DIR) && shasum *
