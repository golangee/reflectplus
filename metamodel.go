package reflectplus

import (
	"encoding/json"
	"fmt"
	"github.com/worldiety/reflectplus/parser"
	"strconv"
)

type Package struct {
	Doc        string       `json:"doc,omitempty"`
	ImportPath string       `json:"importPath,omitempty"`
	Name       string       `json:"name,omitempty"`
	Packages   []*Package   `json:"packages,omitempty"`
	Interfaces []*Interface `json:"interfaces,omitempty"`
	Structs    []*Struct    `json:"structs,omitempty"`
	Funcs      []Method     `json:"funcs,omitempty"`
}

func (p *Package) String() string {
	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (p *Package) VisitPackages(f func(pkg Package) bool) {
	for _, c := range p.Packages {
		if !f(*c) {
			return
		}
		c.VisitPackages(f)
	}
	return
}

func (p *Package) AllInterfaces() []Interface {
	var res []Interface
	for _, iface := range p.Interfaces {
		res = append(res, *iface)
	}
	for _, pkg := range p.Packages {
		res = append(res, pkg.AllInterfaces()...)
	}
	return res
}

// An Annotation is actually an @-prefixed-named json object one-liner
type Annotation struct {
	Doc    string
	Text   string
	Name   string
	Values map[string]interface{}
}

func (a Annotation) AsString(key string) string {
	if a.Values == nil {
		return ""
	}

	v := a.Values[key]
	if v == nil {
		return ""
	}

	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprintf("%v", v)
}

func (a Annotation) AsFloat(key string) float64 {
	if a.Values == nil {
		return 0
	}

	v := a.Values[key]
	if v == nil {
		return 0
	}

	if s, ok := v.(float64); ok {
		return s
	}

	s := a.AsString(key)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (a Annotation) AsBool(key string) bool {
	if a.Values == nil {
		return false
	}

	v := a.Values[key]
	if v == nil {
		return false
	}

	if s, ok := v.(bool); ok {
		return s
	}

	s := a.AsString(key)
	b, _ := strconv.ParseBool(s)
	return b
}

type Interface struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	ImportPath  string       `json:"importPath,omitempty"`
	Name        string       `json:"name,omitempty"`
	Methods     []*Method    `json:"methods,omitempty"`
	Pos         Pos          `json:"pos,omitempty"`
}

func (p *Interface) String() string {
	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Method struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	Receiver    *Param       `json:"receiver,omitempty"` // optional receiver, if this is actually a struct method not a function
	Name        string       `json:"name,omitempty"`
	Params      []*Param     `json:"params,omitempty"`
	Returns     []*Param     `json:"returns,omitempty"`
	Pos         Pos          `json:"pos,omitempty"`
}

type Param struct {
	Doc  string   `json:"doc,omitempty"`
	Name string   `json:"name,omitempty"`
	Type TypeDecl `json:"type,omitempty"`
}

func ParseMetaModel(pkg *parser.Package) (*Package, error) {
	res := &Package{
		ImportPath: pkg.ImportPath(),
		Name:       pkg.Name(),
		Doc:        pkg.Doc().Doc,
	}

	ifaces, err := parseInterfaces(pkg)
	if err != nil {
		return nil, err
	}
	res.Interfaces = ifaces
	_ = ifaces

	structs, err := parseStructs(pkg)
	if err != nil {
		return nil, err
	}
	res.Structs = structs
	_ = structs

	pkgFuncs, err := parsePackageFuncs(pkg)
	if err != nil {
		return nil, err
	}
	res.Funcs = pkgFuncs
	_ = pkgFuncs

	for _, p := range pkg.Packages() {
		pkg, err := ParseMetaModel(p)
		if err != nil {
			return nil, err
		}
		res.Packages = append(res.Packages, pkg)
	}

	return res, nil
}

// A TypeDecl (TypeDeclaration) refers to a type definition somewhere else. A declaration may contain other type
// parameter for generics (currently only slices, maps and channels), which itself may be generic. Also in a
// parameter definition variable (ellipsis) is allowed. What makes it even more complex are length attributes for arrays
// and an variable amount of pointer indirection (stars).
type TypeDecl struct {
	ImportPath string     `json:"importPath,omitempty"`
	Identifier string     `json:"identifier,omitempty"` // slices and arrays are [], maps are map, look at the type Params for details. func is a hard one and is describes in Func
	Stars      int        `json:"stars,omitempty"`
	Var        bool       `json:"var,omitempty"`
	Params     []TypeDecl `json:"params,omitempty"` // generics: currently only slices [], arrays [x] and maps map[a]b are supported
	Length     int        `json:"length,omitempty"` // parsed array length or -1 for a slice, 0 if not applicable
	Func       *Method    `json:"func,omitempty"`   // only non-nil if identifier is "func"
}

type Struct struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	ImportPath  string       `json:"importPath,omitempty"`
	Name        string       `json:"name,omitempty"`
	Fields      []Field      `json:"fields,omitempty"`
	Methods     []*Method    `json:"methods,omitempty"`   // Methods with a receiver
	Factories   []*Method    `json:"factories,omitempty"` // factory methods, returning the struct
	Pos         Pos          `json:"pos,omitempty"`
}

type Field struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	Name        string
	Type        TypeDecl          `json:"type,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Pos         Pos               `json:"pos,omitempty"`
}

type Pos struct {
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
}
