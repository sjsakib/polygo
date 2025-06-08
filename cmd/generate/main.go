package main

import (
	"flag"
	"log"

	"github.com/sjsakib/polygo/generator"
	"github.com/sjsakib/polygo/languages"
)

func main() {
	path := flag.String("path", ".", "path to the go package")

	flag.Parse()

	generator := generator.NewGenerator(*path, languages.NewTypescriptGenerator())

	err := generator.Generate()
	if err != nil {
		log.Fatalf("error generating: %s", err)
	}
}
