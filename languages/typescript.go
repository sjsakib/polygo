package languages

import (
	"fmt"
	"go/ast"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sjsakib/polygo/parser"
	"github.com/sjsakib/polygo/utils"
)

type TypescriptGenerator struct {
	libBuilder   *strings.Builder
	typesBuilder *strings.Builder
	pkgs         []*parser.Package
}

func NewTypescriptGenerator() *TypescriptGenerator {
	return &TypescriptGenerator{
		libBuilder:   &strings.Builder{},
		typesBuilder: &strings.Builder{},
		pkgs:         make([]*parser.Package, 0),
	}
}

func (g *TypescriptGenerator) SetPackages(pkgs []*parser.Package) {
	g.pkgs = pkgs
}

func (g *TypescriptGenerator) GetOutputDirname() string {
	return "ts"
}

func (g *TypescriptGenerator) GenerateOutputFiles() ([]*OutputFile, error) {

	g.typesBuilder.WriteString("declare module \"@polygo/example\" {\n")

	g.generateBootstrap()

	for _, pkg := range g.pkgs {
		log.Printf("generating typescript interface for: %s", pkg.Name)
		for _, typeSpec := range pkg.TypesDefs {
			g.generateType(typeSpec)
		}

		log.Printf("Generating functions for: %s", pkg.Name)
		for _, function := range pkg.Functions {
			log.Printf("Generating function: %s", function.Name.Name)
			g.generateFunctions(function)
		}
	}
	g.typesBuilder.WriteString("}\n")

	return []*OutputFile{
		{
			Path:    "types.d.ts",
			Content: g.typesBuilder.String(),
		},
		{
			Path:    "lib.js",
			Content: g.libBuilder.String(),
		},
		{
			Path:    "package.json",
			Content: g.generatePackageJson(),
		},
	}, nil
}

func (g *TypescriptGenerator) generateType(typeSpec *ast.TypeSpec) {
	log.Printf("generating typescript interface for: %s", typeSpec.Name.Name)
	strkt, ok := typeSpec.Type.(*ast.StructType)

	if !ok {
		log.Printf("type %s is not a struct", typeSpec.Name.Name)
		return
	}

	g.typesBuilder.WriteString("export interface ")
	g.typesBuilder.WriteString(typeSpec.Name.Name)
	g.typesBuilder.WriteString(" {\n")
	for _, field := range strkt.Fields.List {
		name := field.Names[0].Name
		jsonTags := parseJsonTag(field)
		optional := false
		isNullable := isPointer(&field.Type)
		if jsonTags != nil {
			name = jsonTags.FieldName
			optional = jsonTags.Omitempty
		}
		typ := field.Type
		g.typesBuilder.WriteString(fmt.Sprintf("  %s", name))
		if optional {
			g.typesBuilder.WriteString("?")
		}
		g.typesBuilder.WriteString(": ")
		g.typesBuilder.WriteString(mapType(&typ))
		if isNullable {
			g.typesBuilder.WriteString(" | null")
		}
		g.typesBuilder.WriteString(";\n")
	}

	g.typesBuilder.WriteString("}\n")

}

func (g *TypescriptGenerator) generateBootstrap() {
	g.libBuilder.WriteString(`import { readFile } from 'fs/promises';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { createRequire } from 'module';
const __dirname = dirname(fileURLToPath(import.meta.url));
export async function goBootstrap() {
	const require = createRequire(import.meta.url);
	const vm = require('vm');
	const goWasm = await readFile(require.resolve('./wasm_exec.js'), 'utf8');
	vm.runInThisContext(goWasm);
	const go = new global.Go();
`)
	g.libBuilder.WriteString("	const wasmBytes = await readFile(`${__dirname}/main.wasm`);\n")

	g.libBuilder.WriteString(`	const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
	go.run(result.instance);
}
`)

	g.typesBuilder.WriteString("export function goBootstrap(): Promise<void>;\n")

}
func (g *TypescriptGenerator) generatePackageJson() string {
	return `
{
    "name": "@polygo/example",
    "version": "1.0.0",
    "main": "lib.js",
    "types": "types.d.ts",
    "type": "module",
    "files": [
        "lib.js",
        "types.d.ts",
        "main.wasm",
        "wasm_exec.js"
    ]
}`

}

func (g *TypescriptGenerator) generateFunctions(function *ast.FuncDecl) {
	log.Printf("generating function: %s", function.Name.Name)
	g.libBuilder.WriteString("export function ")
	g.libBuilder.WriteString(function.Name.Name)
	g.libBuilder.WriteString("(")
	for i, param := range function.Type.Params.List {
		g.libBuilder.WriteString(param.Names[0].Name)
		if i < len(function.Type.Params.List)-1 {
			g.libBuilder.WriteString(", ")
		}
	}
	g.libBuilder.WriteString(") {\n")

	g.libBuilder.WriteString("const [result, err] = global.")
	g.libBuilder.WriteString(function.Name.Name)
	g.libBuilder.WriteString("(")
	for i, param := range function.Type.Params.List {
		g.libBuilder.WriteString(param.Names[0].Name)
		if i < len(function.Type.Params.List)-1 {
			g.libBuilder.WriteString(", ")
		}
	}
	g.libBuilder.WriteString(");\n")
	g.libBuilder.WriteString("if (err) {\n")
	g.libBuilder.WriteString("throw new Error(err);\n")
	g.libBuilder.WriteString("}\n")
	g.libBuilder.WriteString("return result;\n")
	g.libBuilder.WriteString("}\n")

	g.typesBuilder.WriteString("export function ")
	g.typesBuilder.WriteString(function.Name.Name)
	g.typesBuilder.WriteString("(")
	for i, param := range function.Type.Params.List {
		g.typesBuilder.WriteString(param.Names[0].Name)
		g.typesBuilder.WriteString(": ")
		g.typesBuilder.WriteString(mapType(&param.Type))
		if i < len(function.Type.Params.List)-1 {
			g.typesBuilder.WriteString(", ")
		}
	}
	g.typesBuilder.WriteString("): ")
	if function.Type.Results != nil && len(function.Type.Results.List) > 0 {
		g.typesBuilder.WriteString(mapType(&function.Type.Results.List[0].Type))
	} else {
		g.typesBuilder.WriteString("void")
	}
	g.typesBuilder.WriteString(";\n")

}

func (g *TypescriptGenerator) GenerateSourceFiles() ([]*OutputFile, error) {

	wrapperBuilder := &strings.Builder{}
	files := make([]*OutputFile, 0)

	for _, pkg := range g.pkgs {
		log.Printf("generating function wrappers for: %s", pkg.Name)

		wrapperBuilder.WriteString("package ")
		wrapperBuilder.WriteString(pkg.Name)
		wrapperBuilder.WriteString("\n")
		wrapperBuilder.WriteString("import \"encoding/json\"\n")
		wrapperBuilder.WriteString("import \"syscall/js\"\n")
		wrapperBuilder.WriteString("\n")

		mainBuilder := &strings.Builder{}

		mainBuilder.WriteString("func main() {\n")

		for _, function := range pkg.Functions {
			if function.Name.Name == "main" {
				continue
			}
			wrapperBuilder.WriteString("func ")
			wrapperBuilder.WriteString(function.Name.Name + "Wrapper")
			wrapperBuilder.WriteString("(")

			wrapperBuilder.WriteString("this js.Value, p []js.Value) any {\n")
			for i, param := range function.Type.Params.List {
				wrapperBuilder.WriteString(fmt.Sprintf("var param%d", i))
				wrapperBuilder.WriteString(typeToString(&param.Type))
				wrapperBuilder.WriteString("\n")
				wrapperBuilder.WriteString(fmt.Sprintf("err := json.Unmarshal([]byte(js.Global().Get(\"JSON\").Call(\"stringify\", p[%d]).String()), &param%d)\n", i, i))
				wrapperBuilder.WriteString("if err != nil {\n")
				wrapperBuilder.WriteString("return js.ValueOf([]any{js.Null(), \"failed to parse argument: \" + err.Error()})\n")
				wrapperBuilder.WriteString("}\n")

			}
			wrapperBuilder.WriteString("result := ")
			wrapperBuilder.WriteString(function.Name.Name)
			wrapperBuilder.WriteString("(")
			for i := range function.Type.Params.List {
				wrapperBuilder.WriteString(fmt.Sprintf("param%d", i))
				if i < len(function.Type.Params.List)-1 {
					wrapperBuilder.WriteString(", ")
				}
			}
			wrapperBuilder.WriteString(")\n")
			wrapperBuilder.WriteString("return []any{result, js.Null()}\n")
			wrapperBuilder.WriteString("}\n")

			mainBuilder.WriteString(fmt.Sprintf("js.Global().Set(\"%s\", js.FuncOf(%sWrapper))\n", function.Name.Name, function.Name.Name))
		}

		mainBuilder.WriteString("select {}\n")

		mainBuilder.WriteString("}\n")

		files = append(files, &OutputFile{
			Path:    "lib.go",
			Content: wrapperBuilder.String() + "\n" + mainBuilder.String(),
		})
	}
	return files, nil
}

func (g *TypescriptGenerator) Compile(outputPath string, tmpPath string) error {

	cmd := exec.Command("go", "build", "-o", "main.wasm")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Dir = tmpPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("error compiling go code: %s", err)
		return err
	}

	utils.CopyFile(filepath.Join(tmpPath, "main.wasm"), filepath.Join(outputPath, "main.wasm"))

	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		log.Printf("error getting GOROOT: %s", err)
		return err
	}

	goRoot := strings.TrimSpace(string(out))

	wasmExecPath := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")
	wasmExecContent, err := os.ReadFile(wasmExecPath)
	if err != nil {
		log.Printf("error reading wasm_exec.js: %s", err)
		return err
	}
	os.WriteFile(filepath.Join(outputPath, "wasm_exec.js"), wasmExecContent, 0644)

	return nil
}

func typeToString(typ *ast.Expr) string {
	switch x := (*typ).(type) {
	case *ast.Ident:
		return x.Name
	case *ast.ArrayType:
		return "[]" + typeToString(&x.Elt)
	case *ast.MapType:
		return "unimplemented"
	case *ast.StructType:
		return "any"
	case *ast.StarExpr:
		return "*" + typeToString(&x.X)
	}
	return ""
}

func mapType(typ *ast.Expr) string {
	switch x := (*typ).(type) {
	case *ast.Ident:
		return mapBasicType(x.Name)
	case *ast.ArrayType:
		return mapType(&x.Elt) + "[]"
	case *ast.MapType:
		return "Record<string, " + mapType(&x.Value) + ">"
	case *ast.StructType:
		fmt.Println("struct type found")
		return "any"
	case *ast.StarExpr:
		return mapType(&x.X)
	}

	return ""
}

func isPointer(typ *ast.Expr) bool {
	switch (*typ).(type) {
	case *ast.StarExpr:
		return true
	}
	return false
}

func mapBasicType(typName string) string {
	switch typName {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64", "float", "float32", "float64":
		return "number"
	case "char":
		return "string"
	}

	return typName
}

type jsonTag struct {
	FieldName string
	Omitempty bool
}

func parseJsonTag(field *ast.Field) *jsonTag {
	if field.Tag == nil {
		return nil
	}

	tagStr := strings.ReplaceAll(field.Tag.Value, "`", "")

	json := reflect.StructTag(tagStr).Get("json")

	if json == "" {
		return nil
	}

	parts := strings.Split(json, ",")

	return &jsonTag{
		FieldName: parts[0],
		Omitempty: strings.Contains(json, "omitempty"),
	}
}
