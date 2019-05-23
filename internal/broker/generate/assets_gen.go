// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {

	var fs http.FileSystem = http.Dir("./internal/broker/assets")

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:    "./internal/broker/assets.go",
		PackageName: "broker",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
