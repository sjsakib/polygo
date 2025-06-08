package generator

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sjsakib/polygo/languages"
	"github.com/sjsakib/polygo/parser"
	"github.com/sjsakib/polygo/utils"
)

type Generator struct {
	parser          *parser.Parser
	path            string
	targetGenerator languages.LangGenerator
}

func NewGenerator(path string, targetGenerator languages.LangGenerator) *Generator {
	return &Generator{
		parser:          parser.NewParser(path),
		path:            path,
		targetGenerator: targetGenerator,
	}
}

func (g *Generator) Generate() error {
	log.Printf("Generating from: %s", g.path)
	g.parser.Parse()

	packages := g.parser.GetPackages()

	absPath, err := filepath.Abs(g.path)
	if err != nil {
		log.Printf("error getting absolute path: %s", err)
		return err
	}

	g.path = absPath

	outputPath := filepath.Join(absPath, "output", g.targetGenerator.GetOutputDirname())

	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		log.Printf("error creating output directory: %s", err)
		return err
	}

	g.targetGenerator.SetPackages(packages)

	files, err := g.targetGenerator.GenerateOutputFiles()
	if err != nil {
		log.Printf("error generating files: %s", err)
		return err
	}

	for _, file := range files {
		os.WriteFile(filepath.Join(outputPath, file.Path), []byte(file.Content), 0644)
	}

	tmpPath := filepath.Join(g.path, ".tmp")

	err = os.MkdirAll(tmpPath, 0755)
	if err != nil {
		log.Printf("error creating target directory: %s", err)
		return err
	}

	err = utils.CopyDirectoryWithFiles(g.path, tmpPath, []string{".tmp", "output", ".git"})

	if err != nil {
		log.Printf("error copying files: %s", err)
		return err
	}

	files, err = g.targetGenerator.GenerateSourceFiles()
	if err != nil {
		log.Printf("error generating function wrappers: %s", err)
		return err
	}

	for _, file := range files {
		os.WriteFile(filepath.Join(tmpPath, file.Path), []byte(file.Content), 0644)
	}

	err = g.targetGenerator.Compile(outputPath, tmpPath)
	if err != nil {
		log.Printf("error compiling: %s", err)
		return err
	}

	os.RemoveAll(tmpPath)

	return nil
}
