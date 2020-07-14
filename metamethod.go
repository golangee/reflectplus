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
	"github.com/golangee/reflectplus/parser"
	"go/ast"
	"reflect"
	"strconv"
)

func parsePackageFuncs(pkg *parser.Package) ([]Method, error) {
	var res []Method
	var err error
	for _, file := range pkg.Files() {
		ast.Inspect(file.Node(), func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.FuncDecl:
				if t.Recv != nil || !t.Name.IsExported() {
					return true
				}
				doc := pkg.FuncDoc(t.Name.Name)
				method, e := newMethod(file, doc.Doc, t.Name.Name, t.Type)
				if e != nil {
					err = e
					return false
				}
				res = append(res, method)
			}
			return true
		})
	}
	return res, err
}

func parseMethods(ctx *parser.File, methods []*ast.Field) ([]Method, error) {
	var res []Method
	for _, m := range methods {
		if len(m.Names) == 0 || !m.Names[0].IsExported() {
			continue
		}
		switch f := m.Type.(type) {
		case *ast.FuncType:
			method, err := newMethod(ctx, m.Doc.Text(), m.Names[0].Name, f)
			if err != nil {
				return nil, err
			}
			res = append(res, method)
		}

	}
	return res, nil
}

func newMethod(ctx *parser.File, doc string, name string, ftype *ast.FuncType) (Method, error) {
	method := parseMethod(ctx, ftype)
	method.Name = name
	method.Doc = doc

	annotations, e := parser.ParseAnnotations(method.Doc)
	if e != nil {
		return method, newParseErr(ctx, ftype.Pos(), e)
	}
	method.Annotations = wrapAnnotations(method.Pos, annotations)
	return method, nil
}

func parseMethod(ctx *parser.File, f *ast.FuncType) Method {
	method := Method{}

	for _, p := range f.Params.List {
		typeDec := typeDeclOf(ctx, p.Type)

		// go allows anonymous parameters...
		if len(p.Names) == 0 {
			param := Param{
				Doc:  p.Doc.Text(),
				Type: typeDec,
			}
			method.Params = append(method.Params, param)
		}

		// ... and multiple names per type declaration
		for _, name := range p.Names {
			param := Param{
				Doc:  p.Doc.Text(),
				Type: typeDec,
				Name: name.Name,
			}
			method.Params = append(method.Params, param)
		}
	}

	if f.Results != nil {
		for _, p := range f.Results.List {
			result := Param{
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

	method.Pos = posOf(ctx, f.Pos())

	return method
}

func typeDeclOf(ctx *parser.File, exp ast.Expr) TypeDecl {
	switch t := exp.(type) {
	case *ast.Ident:
		return TypeDecl{Identifier: t.Name, ImportPath: ctx.ResolveIdentifierImportName(t.Name)}
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
	case *ast.FuncType:
		method := parseMethod(ctx, t)
		return TypeDecl{Identifier: "func", Func: &method}
	case *ast.StructType:
		return TypeDecl{Identifier: "struct{}"}
	}

	panic(reflect.TypeOf(exp))
}
