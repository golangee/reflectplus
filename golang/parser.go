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

package golang

import (
	"fmt"
	"github.com/golangee/reflectplus/internal/annotation"
	"github.com/golangee/reflectplus/internal/tag"
	"github.com/golangee/reflectplus/meta"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type parseCtx struct {
	fset  *token.FileSet
	files []*ast.File
}

func NewProject(opts Options) (*Project, error) {
	fmt.Println("dir:", opts.Dir)
	fmt.Println("patterns:", opts.Patterns)
	table := meta.NewTable()
	parseCtx := &parseCtx{}
	mtx := sync.Mutex{}
	cfg := &packages.Config{
		Mode:    packages.LoadAllSyntax | packages.NeedModule,
		Context: nil,
		Logf: func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
			fmt.Println()
		},
		Dir:        opts.Dir,
		Env:        nil,
		BuildFlags: nil,
		Fset:       token.NewFileSet(),
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			//TODO remove func body ast to speed up parsing
			const mode = parser.AllErrors | parser.ParseComments
			file, err := parser.ParseFile(fset, filename, src, mode)
			if err != nil {
				return nil, err
			}

			mtx.Lock()
			defer mtx.Unlock()

			parseCtx.files = append(parseCtx.files, file)
			return file, nil
		},
		Tests:   false,
		Overlay: nil,
	}
	parseCtx.fset = cfg.Fset

	//pkgs, err := packages.Load(cfg, "github.com/worldiety/mercurius/...")
	pkgs, err := packages.Load(cfg, opts.Patterns...)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		/*for expr, tv := range pkg.TypesInfo.Declarations{
			posn := cfg.Fset.Position(expr.Pos())
			tvstr := tv.Type.String()
			if tv.Value != nil {
				tvstr += " = " + tv.Value.String()
			}
			if mode(tv)!="type"{
			//	continue
			}
			if !strings.Contains(tvstr,"ProviderType"){
				continue
			}
			// line:col | expr | mode : type = value
			fmt.Fprintf(os.Stdout, "%4d:%4d | %-30s | %-7s : %s\n",
				posn.Line, posn.Column, exprString(cfg.Fset, expr),
				mode(tv), tvstr)

		}*/

		for a, b := range pkg.TypesInfo.Defs {
			/*if !strings.Contains(a.Name, "ProviderType") {
				continue
			}*/
			if a.Obj != nil {
				if _, ok := a.Obj.Decl.(*ast.TypeSpec); ok {
					//addType(table, cfg.Fset, b)
					_, err := putType(table, parseCtx, b.Type())
					if err != nil {
						return nil, err
					}
				}
			}

		}

	}

	prj := &Project{table: table}
	prj.importTable = prj.table.CreateImportTable()

	return prj, nil
}

func putType(table *meta.Table, fset *parseCtx, typ types.Type) (meta.DeclId, error) {
	switch t := typ.(type) {
	case *types.Named:
		return putNamedType(table, fset, t)
	case *types.Struct:
		return putStruct(table, fset, t)
	case *types.Pointer:
		return putPointer(table, fset, t)
	case *types.Array:
		return putArray(table, fset, t)
	case *types.Basic:
		return putBasic(table, fset, t)
	case *types.Slice:
		return putSlice(table, fset, t)
	case *types.Signature:
		return putSignature(table, fset, t)
	case *types.Map:
		return putMap(table, fset, t)
	case *types.Chan:
		return putChan(table, fset, t)
	case *types.Interface:
		return putInterface(table, fset, t)
	default:
		panic(reflect.TypeOf(t))
	}
}

func putChan(table *meta.Table, fset *parseCtx, obj *types.Chan) (meta.DeclId, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	myDir := meta.ChanDir("")
	switch obj.Dir() {
	case types.SendRecv:
		myDir = meta.SendRecv
	case types.RecvOnly:
		myDir = meta.RecvOnly
	case types.SendOnly:
		myDir = meta.SendOnly
	default:
		panic("invalid chan dir:" + strconv.Itoa(int(obj.Dir())))
	}

	res := &meta.Channel{
		ChanDir: myDir,
		DeclId:  tQual,
	}

	id := meta.NewDeclId().Put(res.DeclId, obj.Dir()).Finish()
	table.PutDeclaration(id, meta.Type{
		Channel: res,
	})

	return id, nil
}

func putMap(table *meta.Table, fset *parseCtx, obj *types.Map) (meta.DeclId, error) {
	kQual, err := putType(table, fset, obj.Key())
	if err != nil {
		return "", err
	}

	vQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Map{
		Key:   kQual,
		Value: vQual,
	}

	q := meta.NewDeclId().Put("map", res.Key, res.Value).Finish()
	table.PutDeclaration(q, meta.Type{
		Map: res,
	})

	return q, nil
}

func putSignature(table *meta.Table, fset *parseCtx, obj *types.Signature) (meta.DeclId, error) {
	res := &meta.Signature{}

	builder := meta.NewDeclId()
	builder.Put("signature")
	for i := 0; i < obj.Params().Len(); i++ {
		param := obj.Params().At(i)
		pQual, err := putType(table, fset, param.Type())
		if err != nil {
			return "", err
		}

		p := meta.Param{
			Name:   param.Name(),
			DeclId: pQual,
		}
		res.Params = append(res.Params, p)

		builder.Put(p.Name, p.DeclId)
	}

	for i := 0; i < obj.Results().Len(); i++ {
		param := obj.Results().At(i)
		pQual, err := putType(table, fset, param.Type())
		if err != nil {
			return "", err
		}

		p := meta.Param{
			Name:   param.Name(),
			DeclId: pQual,
		}
		res.Results = append(res.Results, p)

		builder.Put(p.Name, p.DeclId)
	}

	res.Variadic = obj.Variadic()
	builder.Put(res.Variadic)

	if obj.Recv() != nil {
		rQual, err := putType(table, fset, obj.Recv().Type())
		if err != nil {
			return "", err
		}

		res.Receiver = &meta.Param{
			Name:   obj.Recv().Name(),
			DeclId: rQual,
		}

		builder.Put(res.Receiver.Name)
		builder.Put(res.Receiver.DeclId)
	}

	q := builder.Finish()

	table.PutDeclaration(q, meta.Type{
		Signature: res,
	})

	return q, nil
}

func putFunc(table *meta.Table, fset *parseCtx, obj *types.Func) (meta.DeclId, error) {
	pos := fset.fset.Position(obj.Pos())

	pkgImportPath := ""
	pkgName := ""

	if obj.Pkg() != nil {
		pkgImportPath = obj.Pkg().Path()
		pkgName = obj.Pkg().Name()
	}

	qualifier := meta.NewDeclId().Put("func", pkgImportPath, pkgName, obj.Name()).Finish()

	if table.HasDeclaration(qualifier) {
		return qualifier, nil
	}

	loc := meta.NewLocation(pos.Filename, pos.Line, pos.Column)

	s := findTypeComment(fset, obj.Pos())
	annotations, err := annotation.Parse(s)
	if err != nil {
		return "", fmt.Errorf("%s: %w", loc, err)
	}

	uQual, err := putType(table, fset, obj.Type().Underlying())
	if err != nil {
		return "", err
	}

	table.PutNamedDeclaration(pkgImportPath, pkgName, qualifier, &meta.Named{
		Location:    loc,
		Doc:         s,
		Annotations: wrapAnnotations(loc, annotations),
		Underlying:  uQual,
		Name:        obj.Name(),
	})

	return qualifier, nil
}

func putInterface(table *meta.Table, fset *parseCtx, obj *types.Interface) (meta.DeclId, error) {
	res := &meta.Interface{}

	builder := meta.NewDeclId()
	builder.Put("interface")
	for i := 0; i < obj.NumMethods(); i++ {
		methodQualifier, err := putFunc(table, fset, obj.Method(i))
		if err != nil {
			return "", err
		}

		res.AllMethods = append(res.AllMethods, methodQualifier)
		builder.Put(methodQualifier)
	}

	for i := 0; i < obj.NumEmbeddeds(); i++ {
		typeQualifier, err := putType(table, fset, obj.EmbeddedType(i))
		if err != nil {
			return "", err
		}
		res.Embeddeds = append(res.Embeddeds, typeQualifier)
		builder.Put(typeQualifier)
	}

	q := builder.Finish()

	table.PutDeclaration(q, meta.Type{
		Interface: res,
	})

	return q, nil
}

func putSlice(table *meta.Table, fset *parseCtx, obj *types.Slice) (meta.DeclId, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Slice{
		DeclId: tQual,
	}

	q := meta.NewDeclId().Put("slice", res.DeclId).Finish()

	table.PutDeclaration(q, meta.Type{
		Slice: res,
	})

	return q, nil
}

func putBasic(table *meta.Table, fset *parseCtx, obj *types.Basic) (meta.DeclId, error) {
	myKind := meta.BasicKind("")
	switch obj.Kind() {
	case types.Bool:
		myKind = meta.Bool
	case types.Int:
		myKind = meta.Int
	case types.Int8:
		myKind = meta.Int8
	case types.Int16:
		myKind = meta.Int16
	case types.Int32:
		myKind = meta.Int32
	case types.Int64:
		myKind = meta.Int64
	case types.Uint:
		myKind = meta.Uint
	case types.Uint8:
		myKind = meta.Uint8
	case types.Uint16:
		myKind = meta.Uint16
	case types.Uint32:
		myKind = meta.Uint32
	case types.Uint64:
		myKind = meta.Uint64
	case types.Uintptr:
		myKind = meta.Uintptr
	case types.Float32:
		myKind = meta.Float32
	case types.Float64:
		myKind = meta.Float64
	case types.Complex64:
		myKind = meta.Complex64
	case types.Complex128:
		myKind = meta.Complex128
	case types.String:
		myKind = meta.String
	case types.UnsafePointer:
		myKind = meta.UnsafePointer
	default:
		panic("not implemented: basic type " + strconv.Itoa(int(obj.Kind())))

	}

	kind := myKind
	did := meta.NewDeclId().Put("basic", kind.String()).Finish()
	if table.HasDeclaration(did) {
		return did, nil
	}
	res := &meta.Basic{Kind: kind}

	table.PutDeclaration(did, meta.Type{
		Basic: res,
	})

	return did, nil
}

func putArray(table *meta.Table, fset *parseCtx, obj *types.Array) (meta.DeclId, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Array{
		Len:    obj.Len(),
		DeclId: tQual,
	}

	q := meta.NewDeclId().Put("array", int(obj.Len()), res.DeclId).Finish()

	table.PutDeclaration(q, meta.Type{
		Array: res,
	})

	return q, nil
}

func putPointer(table *meta.Table, fset *parseCtx, pointer *types.Pointer) (meta.DeclId, error) {
	baseQual, err := putType(table, fset, pointer.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Pointer{Base: baseQual}

	q := meta.NewDeclId().Put("ptr", res.Base).Finish()
	table.PutDeclaration(q, meta.Type{
		Pointer: res,
	})

	return q, nil
}

func putStruct(table *meta.Table, fset *parseCtx, strct *types.Struct) (meta.DeclId, error) {
	builder := meta.NewDeclId()
	builder.Put("struct")
	res := &meta.Struct{}
	for i := 0; i < strct.NumFields(); i++ {
		// TODO what about the tags? this should be a feature of the named declaration?
		f := strct.Field(i)
		tag := strct.Tag(i)

		pQual, err := putType(table, fset, f.Type())
		if err != nil {
			return "", err
		}
		p := meta.Param{
			Name:   f.Name(),
			DeclId: pQual,
		}
		res.Fields = append(res.Fields, p)
		builder.Put(p.Name, p.DeclId, tag)
	}

	q := builder.Finish()

	table.PutDeclaration(q, meta.Type{
		Struct: res,
	})

	return q, nil
}

// findNode picks the ast node by matching the exact position or returns nil
func findNode(ctx *parseCtx, pos token.Pos) (n ast.Node) {
	for _, f := range ctx.files {
		ast.Inspect(f, func(node ast.Node) bool {
			if node == nil {
				return true
			}

			if node.Pos() == pos {
				n = node
				return false
			}
			return true
		})
	}

	return
}

// findTypeComment searches through the ast.TypeSpec and picks the comment.
// See also https://github.com/golang/go/issues/27477#issuecomment-418563062 for details.
func findTypeComment(ctx *parseCtx, pos token.Pos) string {
	var lastGen *ast.GenDecl
	var lastFunc *ast.FuncDecl
	isTypeSpec := false
	isFuncSpec := false
	actualDoc := ""
	done := false

	for _, f := range ctx.files {
		ast.Inspect(f, func(node ast.Node) bool {
			if node == nil {
				return true
			}

			// this closure is invoked, even if the last TypeSpec check returns false
			if genDecl, ok := node.(*ast.GenDecl); ok && !done {
				lastGen = genDecl
				return true
			}

			if x, ok := node.(*ast.FuncDecl); ok && !done {
				lastFunc = x
				return true
			}

			if node.Pos() == pos && !done {
				done = true
				switch t := node.(type) {
				case *ast.TypeSpec:
					isTypeSpec = true
					actualDoc = t.Doc.Text()
				case *ast.Field:
					actualDoc = t.Doc.Text()
				case *ast.Ident:
					isFuncSpec = true //TODO wrong assumption?
				default:
					panic(reflect.TypeOf(t))
				}
				return false
			}
			return true
		})
	}
	if !done {
		return ""
	}

	s := ""
	if isTypeSpec && lastGen != nil {
		s += lastGen.Doc.Text()
	}

	if isFuncSpec && lastFunc != nil {
		s += lastFunc.Doc.Text()
	}

	s += actualDoc

	return strings.TrimSpace(s)
}

func putNamedType(table *meta.Table, fset *parseCtx, obj *types.Named) (meta.DeclId, error) {

	named := obj.Obj()
	pos := fset.fset.Position(named.Pos())
	pkgImportPath := ""
	pkgName := ""

	if named.Pkg() != nil {
		pkgImportPath = named.Pkg().Path()
		pkgName = named.Pkg().Name()
	}

	qualifier := meta.NewDeclId().Put("func", pkgImportPath, pkgName, named.Name()).Finish()

	if table.HasDeclaration(qualifier) {
		return qualifier, nil
	}

	// fill in some dummy type, to avoid endless recursion
	table.PutDeclaration(qualifier, meta.Type{})

	loc := meta.NewLocation(pos.Filename, pos.Line, pos.Column)

	s := findTypeComment(fset, named.Pos())
	annotations, err := annotation.Parse(s)
	if err != nil {
		return "", fmt.Errorf("%s: %w", loc, err)
	}

	myUnderlyingType, err := putType(table, fset, named.Type().Underlying())
	if err != nil {
		return "", err
	}

	res := &meta.Named{
		Location:    loc,
		Doc:         s,
		Annotations: wrapAnnotations(loc, annotations),
		Underlying:  myUnderlyingType,
		Name:        named.Name(),
	}

	// this is ugly, but the information has been lost. We define, that the first underlying type of a
	// struct is always specific and requires another underlying type without tags, docs etc.
	// TODO currently this only works, if the tags are different
	node := findNode(fset, named.Pos())
	if typeSpec, ok := node.(*ast.TypeSpec); ok {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			if structType.Fields != nil {
				// enrich the field param information
				strct := table.Declarations[myUnderlyingType].Struct
				for _, field := range structType.Fields.List {
					for _, fieldName := range field.Names {
						for i, strctField := range strct.Fields {
							if strctField.Name == fieldName.Name {
								pos := fset.fset.Position(field.Pos())
								loc := meta.NewLocation(pos.Filename, pos.Line, pos.Column)
								strctField.Pos = &loc
								if field.Tag != nil {
									strctField.Tag = field.Tag.Value
									strctField.Tags = tag.Parse(field.Tag.Value)
								}

								strctField.Doc = field.Doc.Text()
								fieldAnnos, err := annotation.Parse(strctField.Doc)
								if err != nil {
									return "", fmt.Errorf("%s: %w", loc, err)
								}
								strctField.Annotations = wrapAnnotations(loc, fieldAnnos)
								// reassign loop value
								strct.Fields[i] = strctField
							}
						}
					}

				}
			}
		}
	}

	for i := 0; i < obj.NumMethods(); i++ {
		method := obj.Method(i)
		mQual, err := putFunc(table, fset, method)
		if err != nil {
			return "", err
		}

		res.Methods = append(res.Methods, mQual)
	}

	table.PutNamedDeclaration(pkgImportPath, pkgName, qualifier, res)

	return qualifier, nil
}

func wrapAnnotations(loc meta.Location, list []annotation.Annotation) []meta.Annotation {
	res := make([]meta.Annotation, 0, len(list))
	for _, a := range list {
		res = append(res, meta.Annotation{
			Pos:    loc,
			Doc:    a.Text,
			Name:   a.Name,
			Values: a.Values,
		})
	}

	return res
}
