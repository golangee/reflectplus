// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reflectplus

import (
	"fmt"
	"github.com/golangee/reflectplus/parser"
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
							Pos:        posOf(file, i.Pos()),
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
						iface.Annotations = wrapAnnotations(iface.Pos, annotations)
						res = append(res, iface)
					}
				}
			}
			return true
		})
	}

	return res, err
}

func posOf(file *parser.File, pos token.Pos) Pos {
	fsetPos := file.Parent().FileSet().Position(pos)
	return Pos{Filename: fsetPos.Filename, Line: fsetPos.Line}
}

func newParseErr(file *parser.File, pos token.Pos, err error) error {
	return fmt.Errorf("%s:%d : %w", file.Filename(), file.Parent().FileSet().Position(pos).Line, err)
}

func wrapAnnotations(pos Pos, annotations []parser.Annotation) []Annotation {
	if len(annotations) > 0 {
		tmp := make([]Annotation, len(annotations))
		for i, a := range annotations {
			tmp[i] = Annotation{
				Doc:    a.Doc,
				Text:   a.Text,
				Name:   a.Name,
				Values: a.Values,
				Pos:    pos,
			}
		}
		return tmp
	}
	return nil
}
