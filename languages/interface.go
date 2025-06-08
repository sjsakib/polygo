package languages

import (
	"github.com/sjsakib/polygo/parser"
)

type OutputFile struct {
	Path    string
	Content string
}

type LangGenerator interface {
	SetPackages(pkgs []*parser.Package)
	GetOutputDirname() string
	GenerateOutputFiles() ([]*OutputFile, error)
	GenerateSourceFiles() ([]*OutputFile, error)
	Compile(outputPath string, tmpPath string) error
}


