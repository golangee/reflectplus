package reflectplus

import (
	"github.com/golangee/reflectplus/parser"
	"go/ast"
)

func parseTypeDef(pkg *parser.Package) ([]TypeDef, error) {
	var res []TypeDef
	var err error
	for _, file := range pkg.Files() {
		ast.Inspect(file.Node(), func(n ast.Node) bool {

			switch t := n.(type) {
			case *ast.TypeSpec:
				if t.Name.IsExported() {
					typeDef := TypeDef{
						Pos: posOf(file, t.Pos()),
						ImportPath: file.Parent().ImportPath(),
					}

					if t.Name != nil {
						typeDef.Name = t.Name.Name
					}

					typeDef.UnderlyingType = typeDeclOf(file, t.Type)
					res = append(res, typeDef)
				}
			}
			return true
		})
	}

	return res, err
}
