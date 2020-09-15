// +build ignore

package main

import (
	"log"

	"github.com/sensepost/gowitness/web"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(web.AssetsLocation, vfsgen.Options{
		Filename:     "assets_vfsdata.go",
		PackageName:  "web",
		VariableName: "Assets",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
