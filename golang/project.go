package golang

import (
	"fmt"
	"github.com/golangee/reflectplus/meta"
	"github.com/golangee/reflectplus/src"
	"reflect"
)

type Project struct {
	table       *meta.Table
	importTable map[meta.DeclId]meta.PkgId
}

func (p *Project) String() string {
	return p.table.String()
}

func (p *Project) ForEachTypeAnnotation(annotationName string, f func(a meta.Annotation, named *meta.Named)) {
	for _, v := range p.table.Declarations {
		if v.Named != nil {
			for _, a := range v.Named.Annotations {
				if a.Name == annotationName {
					f(a, v.Named)
				}
			}
		}
	}
}

func (p *Project) ForEachInterface(f func(pkg *meta.Package, id meta.DeclId, named *meta.Named, iface *meta.Interface)) {
	for _, id := range p.table.DeclIds() {
		v := p.table.Declarations[id]
		if v.Named != nil {
			underlying := p.table.Declarations[v.Named.Underlying]
			if underlying.Interface != nil {
				pkgId := p.importTable[id]
				pkg := p.table.Packages[pkgId]
				f(pkg, id, v.Named, underlying.Interface)
			}
		}
	}
}

func (p *Project) TypeDecl(id meta.DeclId) *src.TypeDecl {
	declaredType := p.table.Declarations[id]

	switch t := declaredType.Kind().(type) {
	case *meta.Basic:
		return src.NewTypeDecl(src.Qualifier(t.Kind.String()))
	case *meta.Named:
		pkg := p.table.Packages[p.importTable[id]]
		return src.NewTypeDecl(src.Qualifier(pkg.Path + "." + t.Name))
	case *meta.Map:
		return src.NewMapDecl(p.TypeDecl(t.Key), p.TypeDecl(t.Value))
	case *meta.Slice:
		return src.NewSliceDecl(p.TypeDecl(t.DeclId))
	case *meta.Channel:
		return src.NewChanDecl(p.TypeDecl(t.DeclId))
	case *meta.Pointer:
		return src.NewPointerDecl(p.TypeDecl(t.Base))
	case *meta.Array:
		return src.NewArrayDecl(t.Len, p.TypeDecl(t.DeclId))
	default:
		panic(reflect.TypeOf(t))
	}
}

type MethodContext struct {
	TypeAnnotations   []meta.Annotation
	TypeImpl          *src.TypeBuilder
	MethodAnnotations []meta.Annotation
	Method            *src.FuncBuilder
}

func (p *Project) Implement(id meta.DeclId, f func(ctx MethodContext)) (*src.TypeBuilder, error) {
	named := p.table.Declarations[id]
	if named.Named == nil {
		return nil, fmt.Errorf(string(id) + " is not a named typed")
	}

	iface := p.table.Declarations[named.Named.Underlying]
	if iface.Interface == nil {
		return nil, fmt.Errorf(string(id) + " is not an interface")
	}

	pkg := p.table.Packages[p.importTable[id]]

	strct := src.NewStruct(named.Named.Name + "Impl")
	strct.SetDoc("... implements the interface " + pkg.Path + "." + named.Named.Name + "\n" + named.Named.Doc)
	for _, methId := range iface.Interface.AllMethods {
		fmt.Println("methodId: ", methId)
		namedMethod := p.table.Declarations[methId]
		if namedMethod.Named == nil {
			panic("method '" + string(methId) + "' must refer to a named signature")
		}

		signature := p.table.Declarations[namedMethod.Named.Underlying]

		if signature.Signature == nil {
			panic("named signature '" + string(namedMethod.Named.Underlying) + "' must refer to a signature")
		}

		method := src.NewFunc(namedMethod.Named.Name).
			SetPointerReceiver(true).
			SetDoc(namedMethod.Named.Doc)

		for _, par := range signature.Signature.Params {
			param := src.NewParameter(par.Name, p.TypeDecl(par.DeclId))
			method.AddParams(param)
		}

		for _, par := range signature.Signature.Results {
			param := src.NewParameter(par.Name, p.TypeDecl(par.DeclId))
			method.AddResults(param)
		}

		strct.AddMethods(method)

		f(MethodContext{
			TypeAnnotations:   named.Named.Annotations,
			TypeImpl:          strct,
			MethodAnnotations: namedMethod.Named.Annotations,
			Method:            method,
		})
	}
	return strct, nil
}
