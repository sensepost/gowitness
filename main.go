package main

import (
	"embed"

	"github.com/sensepost/gowitness/cmd"
)

var (
	//go:embed web/assets/* web/templates/*
	assets embed.FS
)

func main() {
	cmd.Embedded = assets

	cmd.Execute()
}
