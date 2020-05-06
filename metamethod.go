package reflectplus

import (
	"github.com/worldiety/reflectplus/parser"
	"go/ast"
	"reflect"
	"strconv"
)

func parseMethods(ctx *parser.File, methods []*ast.Field) []*Method {
	var res []*Method
	for _, m := range methods {
		if len(m.Names) == 0 || !m.Names[0].IsExported() {
			continue
		}
		method := &Method{
			Name: m.Names[0].Name,
			Doc:  m.Doc.Text(),
		}
		switch f := m.Type.(type) {
		case *ast.FuncType:
			for _, p := range f.Params.List {
				typeDec := typeDeclOf(ctx, p.Type)

				// go allows anonymous parameters...
				if len(p.Names) == 0 {
					param := &Param{
						Doc:  p.Doc.Text(),
						Type: typeDec,
					}
					method.Params = append(method.Params, param)
				}

				// ... and multiple names per type declaration
				for _, name := range p.Names {
					param := &Param{
						Doc:  p.Doc.Text(),
						Type: typeDec,
						Name: name.Name,
					}
					method.Params = append(method.Params, param)
				}
			}

			if f.Results != nil {
				for _, p := range f.Results.List {
					result := &Param{
						Doc: p.Doc.Text(),
					}
					// go allows anonymous and named return parameters
					if len(p.Names) > 0 {
						result.Name = p.Names[0].Name
					}
					result.Type = typeDeclOf(ctx, p.Type)
					method.Returns = append(method.Returns, result)
				}
			}

		}

		res = append(res, method)
	}
	return res
}

func typeDeclOf(ctx *parser.File, exp ast.Expr) TypeDecl {
	switch t := exp.(type) {
	case *ast.Ident:
		return TypeDecl{Identifier: t.Name}
	case *ast.SelectorExpr:
		namedImportPath := t.X.(*ast.Ident).Name
		namedImportPath = ctx.ResolveImportName(namedImportPath)
		return TypeDecl{ImportPath: namedImportPath, Identifier: t.Sel.Name}
	case *ast.StarExpr:
		tmp := typeDeclOf(ctx, t.X)
		tmp.Stars++
		return tmp
	case *ast.Ellipsis:
		tmp := typeDeclOf(ctx, t.Elt)
		tmp.Var = true
		return tmp
	case *ast.InterfaceType:
		return TypeDecl{Identifier: "interface{}"}
	case *ast.ArrayType:
		tmp := typeDeclOf(ctx, t.Elt)
		resolvedLength := -1
		switch l := t.Len.(type) {
		case *ast.BasicLit:
			v, err := strconv.ParseInt(l.Value, 10, 64)
			if err != nil {
				panic(err)
			}
			resolvedLength = int(v)
		case *ast.Ident:
			if l.Obj != nil {
				if valSpec, ok := l.Obj.Decl.(*ast.ValueSpec); ok {
					if len(valSpec.Values) == 1 {
						if basicLit, ok := valSpec.Values[0].(*ast.BasicLit); ok {
							v, err := strconv.ParseInt(basicLit.Value, 10, 64)
							if err != nil {
								panic(err)
							}
							resolvedLength = int(v)
						}
					}
				}
			}
		}
		return TypeDecl{Identifier: "[]", Params: []TypeDecl{tmp}, Length: resolvedLength}
	case *ast.MapType:
		key := typeDeclOf(ctx, t.Key)
		val := typeDeclOf(ctx, t.Value)
		return TypeDecl{Identifier: "map", Params: []TypeDecl{key, val}}
	case *ast.ChanType:
		val := typeDeclOf(ctx, t.Value)
		return TypeDecl{Identifier: "chan", Params: []TypeDecl{val}}
	}
	panic(reflect.TypeOf(exp))
}
