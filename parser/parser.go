package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

type Package struct {
	Name      string
	TypesDefs []*ast.TypeSpec
	Functions []*ast.FuncDecl
}

type Parser struct {
	path     string
	packages []*Package
}

func NewParser(path string) *Parser {
	return &Parser{
		path:     path,
		packages: make([]*Package, 0),
	}
}

func (p *Parser) Parse() {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, p.path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("error parsing directory: %s", err)
		return
	}

	if len(pkgs) == 0 {
		log.Printf("no packages found!")
		return
	}

	for _, pkg := range pkgs {
		log.Printf("found package: %s", pkg.Name)

		parsedPackage := &Package{
			Name: pkg.Name,
		}

		for file := range pkg.Files {
			log.Printf("file: %s", file)
		}

		ast.Inspect(pkg, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			switch n := n.(type) {
			case *ast.FuncDecl:
				log.Printf("found function: %s", n.Name.Name)
				parsedPackage.Functions = append(parsedPackage.Functions, n)
			case *ast.TypeSpec:
				log.Printf("found type: %s", n.Name.Name)
				parsedPackage.TypesDefs = append(parsedPackage.TypesDefs, n)
			}

			return true
		})

		p.packages = append(p.packages, parsedPackage)
	}
}

func (p *Parser) GetPackages() []*Package {
	return p.packages
}
