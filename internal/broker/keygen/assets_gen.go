// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var fs http.FileSystem = http.Dir("./internal/broker/keygen/assets")
	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:    "./internal/broker/keygen/assets.go",
		PackageName: "broker",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
