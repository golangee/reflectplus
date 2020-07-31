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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golangee/reflectplus/internal/annotation"
	"github.com/golangee/reflectplus/internal/mod"
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

type Project struct {
}

type parseCtx struct {
	fset  *token.FileSet
	files []*ast.File
}

func NewProject(opts Options, dir string, mods mod.Modules) (*meta.Table, error) {
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
		Dir:        dir,
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
	pkgs, err := packages.Load(cfg, "github.com/golangee/...")
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		/*for expr, tv := range pkg.TypesInfo.Types{
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

	return table, nil
}

func putType(table *meta.Table, fset *parseCtx, typ types.Type) (meta.Qualifier, error) {

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

func putChan(table *meta.Table, fset *parseCtx, obj *types.Chan) (meta.Qualifier, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Channel{
		ChanDir: meta.ChanDir(obj.Dir()), //TODO better switch case?
		TypeID:  tQual,
	}

	var q meta.Qualifier
	switch obj.Dir() {
	case types.SendRecv:
		q = meta.NewQualifier("", "", "chan["+sanitizeQualifierStr(res.TypeID)+"]")
	case types.SendOnly:
		q = meta.NewQualifier("", "", "chan<-["+sanitizeQualifierStr(res.TypeID)+"]")
	case types.RecvOnly:
		q = meta.NewQualifier("", "", "<-chan["+sanitizeQualifierStr(res.TypeID)+"]")
	default:
		panic(obj.Dir())
	}

	table.PutType(q, meta.Type{
		Channel: res,
	})

	return q, nil
}

func putMap(table *meta.Table, fset *parseCtx, obj *types.Map) (meta.Qualifier, error) {
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

	q := meta.NewQualifier("", "", "map["+sanitizeQualifierStr(res.Key)+"]"+sanitizeQualifierStr(res.Value))
	table.PutType(q, meta.Type{
		Map: res,
	})

	return q, nil
}

func putSignature(table *meta.Table, fset *parseCtx, obj *types.Signature) (meta.Qualifier, error) {
	res := &meta.Signature{}

	hasher := sha256.New()
	for i := 0; i < obj.Params().Len(); i++ {
		param := obj.Params().At(i)
		pQual, err := putType(table, fset, param.Type())
		if err != nil {
			return "", err
		}

		p := meta.Param{
			Name:   param.Name(),
			TypeId: pQual,
		}
		res.Params = append(res.Params, p)

		hasher.Write([]byte(p.Name))
		hasher.Write([]byte(p.TypeId))
	}

	for i := 0; i < obj.Results().Len(); i++ {
		param := obj.Results().At(i)
		pQual, err := putType(table, fset, param.Type())
		if err != nil {
			return "", err
		}

		p := meta.Param{
			Name:   param.Name(),
			TypeId: pQual,
		}
		res.Results = append(res.Results, p)

		hasher.Write([]byte(p.Name))
		hasher.Write([]byte(p.TypeId))
	}

	res.Variadic = obj.Variadic()
	hasher.Write([]byte(strconv.FormatBool(res.Variadic)))

	if obj.Recv() != nil {
		rQual, err := putType(table, fset, obj.Recv().Type())
		if err != nil {
			return "", err
		}

		res.Receiver = &meta.Param{
			Name:   obj.Recv().Name(),
			TypeId: rQual,
		}

		hasher.Write([]byte(res.Receiver.Name))
		hasher.Write([]byte(res.Receiver.TypeId))
	}

	q := meta.NewQualifier("", "", "func-"+hex.EncodeToString(hasher.Sum(nil)[:8]))

	table.PutType(q, meta.Type{
		Signature: res,
	})

	return q, nil
}

func putFunc(table *meta.Table, fset *parseCtx, obj *types.Func) (meta.Qualifier, error) {
	pos := fset.fset.Position(obj.Pos())

	var qualifier meta.Qualifier
	if obj.Pkg() != nil {
		qualifier = meta.NewQualifier(obj.Pkg().Path(), obj.Pkg().Name(), obj.Name())
	} else {
		// package less named types are from universe, like error
		qualifier = meta.NewQualifier("", "", obj.Name())
	}

	if table.HasType(qualifier) {
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

	table.PutType(qualifier, meta.Type{
		Named: &meta.Named{
			Location:    loc,
			Doc:         s,
			Annotations: wrapAnnotations(loc, annotations),
			Underlying:  uQual,
			Name:        obj.Name(),
		},
	})

	return qualifier, nil
}

func putInterface(table *meta.Table, fset *parseCtx, obj *types.Interface) (meta.Qualifier, error) {
	res := &meta.Interface{}

	hasher := sha256.New()
	for i := 0; i < obj.NumMethods(); i++ {
		methodQualifier, err := putFunc(table, fset, obj.Method(i))
		if err != nil {
			return "", err
		}

		res.AllMethods = append(res.AllMethods, methodQualifier)
		hasher.Write([]byte(methodQualifier))
	}

	for i := 0; i < obj.NumEmbeddeds(); i++ {
		typeQualifier, err := putType(table, fset, obj.EmbeddedType(i))
		if err != nil {
			return "", err
		}
		res.Embeddeds = append(res.Embeddeds, typeQualifier)
		hasher.Write([]byte(typeQualifier))
	}

	q := meta.NewQualifier("", "", hex.EncodeToString(hasher.Sum(nil)[:8]))

	table.PutType(q, meta.Type{
		Interface: res,
	})

	return q, nil
}

func putSlice(table *meta.Table, fset *parseCtx, obj *types.Slice) (meta.Qualifier, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Slice{
		TypeId: tQual,
	}

	q := meta.NewQualifier("", "", "[]"+sanitizeQualifierStr(res.TypeId))

	table.PutType(q, meta.Type{
		Slice: res,
	})

	return q, nil
}

func sanitizeQualifierStr(q meta.Qualifier) string {
	return strings.ReplaceAll(q.String(), "|", "_")
}

func putBasic(table *meta.Table, fset *parseCtx, obj *types.Basic) (meta.Qualifier, error) {
	kind := meta.BasicKind(obj.Kind())
	qualifer := meta.NewQualifier("", "", kind.String())
	if table.HasType(qualifer) {
		return qualifer, nil
	}
	res := &meta.Basic{Kind: kind} //TODO better switch/case than to rely on internals

	table.PutType(qualifer, meta.Type{
		Basic: res,
	})

	return qualifer, nil
}

func putArray(table *meta.Table, fset *parseCtx, obj *types.Array) (meta.Qualifier, error) {
	tQual, err := putType(table, fset, obj.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Array{
		Len:    obj.Len(),
		TypeID: tQual,
	}

	q := meta.NewQualifier("", "", "["+strconv.Itoa(int(obj.Len()))+"]"+sanitizeQualifierStr(res.TypeID))

	table.PutType(q, meta.Type{
		Array: res,
	})

	return q, nil
}

func putPointer(table *meta.Table, fset *parseCtx, pointer *types.Pointer) (meta.Qualifier, error) {
	baseQual, err := putType(table, fset, pointer.Elem())
	if err != nil {
		return "", err
	}

	res := &meta.Pointer{Base: baseQual}

	q := meta.NewQualifier("", "", "*"+sanitizeQualifierStr(res.Base))
	table.PutType(q, meta.Type{
		Pointer: res,
	})

	return q, nil
}

func putStruct(table *meta.Table, fset *parseCtx, strct *types.Struct) (meta.Qualifier, error) {
	res := &meta.Struct{}
	hasher := sha256.New()
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
			TypeId: pQual,
		}
		res.Fields = append(res.Fields, p)

		hasher.Write([]byte((p.Name)))
		hasher.Write([]byte((p.TypeId)))
		hasher.Write([]byte(tag))
	}

	q := meta.NewQualifier("", "", hex.EncodeToString(hasher.Sum(nil)[:8]))

	table.PutType(q, meta.Type{
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
	if !done{
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

	return s
}

func putNamedType(table *meta.Table, fset *parseCtx, obj *types.Named) (meta.Qualifier, error) {

	named := obj.Obj()
	pos := fset.fset.Position(named.Pos())
	var qualifier meta.Qualifier
	if named.Pkg() != nil {
		qualifier = meta.NewQualifier(named.Pkg().Path(), named.Pkg().Name(), named.Name())
	} else {
		// package less named types are from universe, like error
		qualifier = meta.NewQualifier("", "", named.Name())
	}

	if table.HasType(qualifier) {
		return qualifier, nil
	}

	// fill in some dummy type, to avoid endless recursion
	table.PutType(qualifier, meta.Type{})

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
				strct := table.Types[myUnderlyingType].Struct
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

	table.PutType(qualifier, meta.Type{
		Named: res,
	})

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
