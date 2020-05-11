package reflectplus

import (
	"github.com/worldiety/reflectplus/parser"
	"go/ast"
	"strconv"
	"strings"
)

func parseStructs(pkg *parser.Package) ([]*Struct, error) {
	var res []*Struct
	var err error
	for _, file := range pkg.Files() {
		ast.Inspect(file.Node(), func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.TypeSpec:
				if t.Name.IsExported() {
					switch i := t.Type.(type) {
					case *ast.StructType:
						strct := &Struct{
							Doc:        file.Parent().TypeDoc(t.Name.Name).Doc,
							ImportPath: file.Parent().ImportPath(),
							Name:       t.Name.Name,
							Pos:        posOf(file, i.Pos()),
						}
						/* it is fine to be incomplete, happens for any type declared outside of the fileset
						if i.Incomplete {
							err = newParseErr(file, t.Pos(), fmt.Errorf("%s.%s is incomplete", strct.ImportPath, strct.Name))
							return false
						}*/
						for _, f := range i.Fields.List {
							if len(f.Names) == 0 {
								// anonymous field embedding
								field, e := parseField(file, f, "")
								if e != nil {
									err = e
									return false
								}
								strct.Fields = append(strct.Fields, field)
							} else {
								for _, n := range f.Names {
									field, e := parseField(file, f, n.Name)
									if e != nil {
										err = e
										return false
									}
									strct.Fields = append(strct.Fields, field)
								}
							}
						}

						docType := pkg.TypeDoc(t.Name.Name)
						for _, m := range docType.Methods {
							method, e := newMethod(file, m.Doc, m.Decl.Name.Name, m.Decl.Type)
							if e != nil {
								err = e
								return false
							}

							recField := m.Decl.Recv.List[0]
							receiver := &Param{
								Type: typeDeclOf(file, recField.Type),
							}

							if len(recField.Names) > 0 {
								receiver.Name = recField.Names[0].Name
							}

							method.Receiver = receiver
							strct.Methods = append(strct.Methods, &method)
						}

						for _, m := range docType.Funcs {
							method, e := newMethod(file, m.Doc, m.Decl.Name.Name, m.Decl.Type)
							if e != nil {
								err = e
								return false
							}
							strct.Factories = append(strct.Factories, &method)
						}

						//iface.Methods, err = parseMethods(file, i.Methods.List)
						if err != nil {
							return false
						}
						annotations, e := parser.ParseAnnotations(strct.Doc)
						if e != nil {
							err = newParseErr(file, t.Pos(), e)
							return false
						}
						strct.Annotations = wrapAnnotations(strct.Pos, annotations)
						res = append(res, strct)
					}
				}
			}
			return true
		})
	}

	return res, err
}

func parseField(ctx *parser.File, f *ast.Field, name string) (Field, error) {
	field := Field{
		Doc:         f.Doc.Text(),
		Annotations: nil,
		Name:        name,
		Type:        typeDeclOf(ctx, f.Type),
		Pos:         posOf(ctx, f.Pos()),
	}

	annotations, e := parser.ParseAnnotations(field.Doc)
	if e != nil {
		return field, newParseErr(ctx, f.Pos(), e)
	}
	field.Annotations = wrapAnnotations(field.Pos, annotations)

	if f.Tag != nil {
		tagStr := f.Tag.Value
		if strings.HasPrefix(tagStr, "`") {
			tagStr = tagStr[1:]
		}

		if strings.HasSuffix(tagStr, "`") {
			tagStr = tagStr[:len(tagStr)-1]
		}
		tagStr = strings.TrimSpace(tagStr)
		tags, err := mapFieldTags(tagStr)
		if err != nil {
			return field, newParseErr(ctx, f.Pos(), err)
		}
		field.Tags = tags
	}
	return field, nil
}

// based on reflect/type.go StructTag.Lookup
func mapFieldTags(tag string) (map[string]string, error) {
	res := make(map[string]string)
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, err
		}

		//dunno why this has a broken name, the original code seems to have this as a defacto-bug
		if strings.HasPrefix(name, ",") {
			name = name[1:]
		}
		res[name] = value
	}
	return res, nil
}
