package main

import (
	"log"

	"github.com/ardanlabs/service/cmd/search/views"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(views.Assets, vfsgen.Options{
		Filename:     "views/assets_vfsdata.go",
		PackageName:  "views",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
