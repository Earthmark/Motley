// +build ignore

package main

import (
	"log"

	"github.com/Earthmark/Motley/server/core"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(core.Assets, vfsgen.Options{
		PackageName:  "client",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
