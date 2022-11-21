package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"

	"github.com/griggsca91/gogenerate/parsegofile/pkg"
)

//go:generate parsegofile InterfaceWeWantToParse

type CustomStruct struct{}

type InterfaceWeWantToParse interface {
	PublicMethod1(namedArg string)
	PublicMethod2(string) string
	PublicMethod3() (string, error)
	PublicMethod4() (CustomStruct, error)
	PublicMethod5() (*CustomStruct, error)
	PublicMethod6() (pkg.CustomStruct, error)
	PublicMethod7() (*pkg.CustomStruct, error)
	privateMethod()
}

type InterfaceWeWantIgnore interface {
	PublicMethod1Ignored()
	PublicMethod2Ignored() string
	PublicMethod3Ignored() (string, error)
	privateMethodIgnored()
}

/**

[X] parse a single file
[X] look for an exported interface by the arg name
[X] get all the exported methods of the found interface
[ ] generate a file that implements the interface with NOOP methods
[ ] NOOP Method is a method that returns a 0 value for all the results
[ ] Organize package into a proper library
[ ] Add tests

*/

func main() {
	interfaceName := os.Args[1]
	noopStructName := fmt.Sprintf("Noop%s", interfaceName)
	file, _ := os.LookupEnv("GOFILE")
	log.Println("Current file", file)
	wd, _ := os.Getwd()
	log.Println("current directory", wd)

	log.Println("Args", os.Args)
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("main", fset, []*ast.File{f}, nil)
	if err != nil {
		log.Fatal(err) // type error
	}

	log.Println(pkg)

	return

	ast.FileExports(f)

	var buf bytes.Buffer

	buf.WriteString("package noop\n\n\n")
	buf.WriteString(fmt.Sprintf("type %s struct {}\n", noopStructName))

	ast.Inspect(f, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.TypeSpec:
			if n.Name.Name != interfaceName {
				return true
			}
			println("name", n.Name.Name)
			log.Printf("expr %T\n", n.Type)

			interfaceType, ok := n.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}

			for _, method := range interfaceType.Methods.List {
				f := method.Type.(*ast.FuncType)
				if len(f.Params.List) > 0 {
					for _, param := range f.Params.List {
						log.Println("Param", param.Type)
						log.Printf("Param Type: %T\n", param.Type)
					}
				}
				if f.Results != nil {
					for _, result := range f.Results.List {
						log.Println("Result", result.Type)
						log.Printf("Result Type: %T\n", result.Type)
						switch t := result.Type.(type) {
						case *ast.Ident:
							log.Println("t", t)
						case *ast.StarExpr:
							log.Println("t", t)
						case *ast.SelectorExpr:
							ident := t.X.(*ast.Ident)
							log.Printf("t selector ecpr %v %T, %s\n", ident.Name, t.X, t.Sel.Name)
						}
						println("========")
					}
				}
				println("========")
			}

			return true
		}
		return true
	})

	os.MkdirAll("noop", 0o777)

	os.WriteFile(fmt.Sprintf("noop/%s.go", strings.ToLower(interfaceName)), buf.Bytes(), 0o777)
}

// GenerateMethod returns a string for a method definition to be inserted into a file
// args will expect the array to be a types in their string representation e.g. []string { "int", "any", "RandomStruct" }
// results will have the same expectation as the any argument
func GenerateMethod(receiverName, methodName string, args []string, results []string) string {
	argsString := strings.Join(args, ", ")
	resultsString := strings.Join(results, ", ")
	if len(results) > 1 {
		resultsString = "(" + resultsString + ")"
	}

	zeroValues := ""

	return fmt.Sprintf("(%s) %s(%s) %s {\n"+
		"return %s\n"+
		"}\n", receiverName, methodName, argsString, resultsString, zeroValues)
}
