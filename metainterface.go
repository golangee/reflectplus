package reflectplus

import (
	"fmt"
	"github.com/worldiety/reflectplus/parser"
	"go/ast"
	"go/token"
)

func parseInterfaces(pkg *parser.Package) ([]*Interface, error) {
	var res []*Interface
	var err error
	for _, file := range pkg.Files() {
		ast.Inspect(file.Node(), func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.TypeSpec:
				if t.Name.IsExported() {
					switch i := t.Type.(type) {
					case *ast.InterfaceType:
						iface := &Interface{
							Doc:        file.Parent().TypeDoc(t.Name.Name).Doc,
							ImportPath: file.Parent().ImportPath(),
							Name:       t.Name.Name,
						}
						/* it is fine to be incomplete, happens for any type declared outside of the fileset
						if i.Incomplete {
							err = newParseErr(file, t.Pos(), fmt.Errorf("%s.%s is incomplete", iface.ImportPath, iface.Name))
							return false
						}*/

						iface.Methods, err = parseMethods(file, i.Methods.List)
						if err != nil {
							return false
						}
						annotations, e := parser.ParseAnnotations(iface.Doc)
						if e != nil {
							err = newParseErr(file, t.Pos(), e)
							return false
						}
						iface.Annotations = wrapAnnotations(annotations)
						res = append(res, iface)
					}
				}
			}
			return true
		})
	}

	return res, err
}

func newParseErr(file *parser.File, pos token.Pos, err error) error {
	return fmt.Errorf("%s:%d : %w", file.Filename(), file.Parent().FileSet().Position(pos).Line, err)
}

func wrapAnnotations(annotations []parser.Annotation) []Annotation {
	if len(annotations) > 0 {
		tmp := make([]Annotation, len(annotations))
		for i, a := range annotations {
			tmp[i] = Annotation(a)
		}
		return tmp
	}
	return nil
}
