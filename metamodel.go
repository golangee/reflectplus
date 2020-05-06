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
}

func (p *Package) String() string {
	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
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
}

type Method struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	Name        string       `json:"name,omitempty"`
	Params      []*Param     `json:"params,omitempty"`
	Returns     []*Param     `json:"returns,omitempty"`
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
	for _, p := range pkg.Packages() {
		pkg, err := ParseMetaModel(p)
		if err != nil {
			return nil, err
		}
		res.Packages = append(res.Packages, pkg)
	}
	return res, nil
}

type TypeDecl struct {
	ImportPath string     `json:"importPath,omitempty"`
	Identifier string     `json:"identifier,omitempty"` // slices and arrays are [], maps are map, look at the type Params for details
	Stars      int        `json:"stars,omitempty"`
	Var        bool       `json:"var,omitempty"`
	Params     []TypeDecl `json:"params,omitempty"` // generics: currently only slices [], arrays [x] and maps map[a]b are supported
	Length     int        `json:"length,omitempty"` // parsed array length or -1 for a slice, 0 if not applicable
}
