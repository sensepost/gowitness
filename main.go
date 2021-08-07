package main

import (
	"embed"

	"github.com/sensepost/gowitness/cmd"
)

var (
	//go:embed web/assets/*
	assets embed.FS
	//go:embed web/templates/*
	templates embed.FS
)

func main() {
	cmd.Assets = assets
	cmd.Templates = templates

	cmd.Execute()
}
