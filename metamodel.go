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
	"encoding/json"
	"fmt"
	"github.com/golangee/reflectplus/parser"
	"math"
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
	TypeDefs   []TypeDef    `json:"typeDefs,omitempty"`
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

// AllStructs returns all defined structs recursively.
func (p *Package) AllStructs() []Struct {
	var res []Struct
	for _, strct := range p.Structs {
		res = append(res, *strct)
	}
	for _, pkg := range p.Packages {
		res = append(res, pkg.AllStructs()...)
	}
	return res
}

// AllTypeDefs returns all defined types recursively.
func (p *Package) AllTypeDefs() []TypeDef {
	var res []TypeDef
	for _, def := range p.TypeDefs {
		res = append(res, def)
	}
	for _, pkg := range p.Packages {
		res = append(res, pkg.AllTypeDefs()...)
	}
	return res
}

// PackageByName searches only the direct children and returns the package or nil if not found.
func (p *Package) PackageByName(name string) *Package {
	for _, c := range p.Packages {
		if c.Name == name {
			return c
		}
	}
	return nil
}

type Annotations []Annotation

// Has returns true if at least one annotation with the given name exists
func (s Annotations) Has(name string) bool {
	for _, a := range s {
		if a.Name == name {
			return true
		}
	}
	return false
}

// FindFirst returns the first matching annotation
func (s Annotations) FindFirst(name string) *Annotation {
	for _, a := range s {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

// FindAll returns all annotations with the given name
func (s Annotations) FindAll(name string) []Annotation {
	var r []Annotation
	for _, a := range s {
		if a.Name == name {
			r = append(r, a)
		}
	}
	return r
}

// MustOne asserts that only
func (s Annotations) MustOne(name string) (*Annotation, error) {
	var r *Annotation
	for _, a := range s {
		if a.Name == name {
			if r != nil {
				return nil, fmt.Errorf("annotation '%s' is ambigous but must be unique", name)
			}
			r = &a
		}
	}
	if r == nil {
		err := fmt.Errorf("expected annotation '%s' not found", name)
		if len(s) > 0 {
			return nil, PositionalError(s[0], err)
		} else {
			return nil, err
		}
	}
	return r, nil
}

// An Annotation is actually an @-prefixed-named json object one-liner
type Annotation struct {
	Doc    string                 `json:"doc,omitempty"`
	Text   string                 `json:"text,omitempty"`
	Name   string                 `json:"name,omitempty"`
	Values map[string]interface{} `json:"values,omitempty"`
	Pos    Pos                    `json:"pos,omitempty"`
}

func (a Annotation) Position() Pos {
	return a.Pos
}

func (a Annotation) OptInt(key string, defaultValue int) int {
	if a.Values == nil {
		return defaultValue
	}
	v := a.Values[key]
	if v == nil {
		return defaultValue
	}

	var val float64
	if s, ok := v.(float64); ok {
		val = s
	} else {
		tmp := fmt.Sprintf("%v", v)
		f, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			return defaultValue
		}
		val = f
	}

	const epsilon = 1e-9
	if _, frac := math.Modf(math.Abs(val)); frac < epsilon || frac > 1.0-epsilon {
		return int(val)
	}
	return defaultValue
}

func (a Annotation) OptString(key string, defaultValue string) string {
	if a.Values == nil {
		return defaultValue
	}

	v := a.Values[key]
	if v == nil {
		return defaultValue
	}

	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprintf("%v", v)
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

// Value returns the value for the "value"-key or the empty string
func (a Annotation) Value() string {
	return a.AsString("value")
}

func (a Annotation) MustAsString(key string) string {
	if a.Values == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	v := a.Values[key]
	if v == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprintf("%v", v)
}

func (a Annotation) MustAsFloat(key string) float64 {
	if a.Values == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	v := a.Values[key]
	if v == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	if s, ok := v.(float64); ok {
		return s
	}

	s := a.AsString(key)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(a.Pos.ideString() + " value of '" + key + "' incompatible: " + err.Error())
	}
	return f
}

func (a Annotation) MustAsBool(key string) bool {
	if a.Values == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	v := a.Values[key]
	if v == nil {
		panic(a.Pos.ideString() + " key '" + key + "' not found")
	}

	if s, ok := v.(bool); ok {
		return s
	}

	s := a.AsString(key)
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(a.Pos.ideString() + " value of '" + key + "' incompatible: " + err.Error())
	}
	return b
}

type Interface struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	ImportPath  string       `json:"importPath,omitempty"`
	Name        string       `json:"name,omitempty"`
	Methods     []Method     `json:"methods,omitempty"`
	Pos         Pos          `json:"pos,omitempty"`
}

func (p Interface) GetAnnotations() Annotations {
	return p.Annotations
}

func (p Interface) String() string {
	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (p Interface) Position() Pos {
	return p.Pos
}

type Method struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	Receiver    *Param       `json:"receiver,omitempty"` // optional receiver, if this is actually a struct method not a function
	Name        string       `json:"name,omitempty"`
	Params      []Param      `json:"params,omitempty"`
	Returns     []Param      `json:"returns,omitempty"`
	Pos         Pos          `json:"pos,omitempty"`
}

func (m Method) Position() Pos {
	return m.Pos
}

func (m Method) ParamByName(n string) *Param {
	for _, p := range m.Params {
		if p.Name == n {
			return &p
		}
	}
	return nil
}

func (m Method) ParamAndIndexByName(n string) (*Param, int) {
	for i, p := range m.Params {
		if p.Name == n {
			return &p, i
		}
	}
	return nil, -1
}

func (m Method) GetAnnotations() Annotations {
	return m.Annotations
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

	typeDefs, err := parseTypeDef(pkg)
	if err != nil {
		return nil, err
	}

	res.TypeDefs = typeDefs

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
	ImportPath string `json:"importPath,omitempty"`

	// slices and arrays are [], maps are map, look at the type Params for details. func is a hard one and is describes in Func
	Identifier string `json:"identifier,omitempty"`
	Stars      int    `json:"stars,omitempty"`
	Var        bool   `json:"var,omitempty"`

	// generics: currently only slices [], arrays [x] and maps map[a]b are supported
	Params []TypeDecl `json:"params,omitempty"`
	Length int        `json:"length,omitempty"` // parsed array length or -1 for a slice, 0 if not applicable
	Func   *Method    `json:"func,omitempty"`   // only non-nil if identifier is "func"
}

type Struct struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
	ImportPath  string       `json:"importPath,omitempty"`
	Name        string       `json:"name,omitempty"`
	Fields      []Field      `json:"fields,omitempty"`
	Methods     []Method     `json:"methods,omitempty"`   // Methods with a receiver
	Factories   []Method     `json:"factories,omitempty"` // factory methods, returning the struct
	Pos         Pos          `json:"pos,omitempty"`
}

func (s Struct) GetAnnotations() Annotations {
	return s.Annotations
}

func (s Struct) Position() Pos {
	return s.Pos
}

type Field struct {
	Doc         string            `json:"doc,omitempty"`
	Annotations []Annotation      `json:"annotations,omitempty"`
	Name        string            `json:"name,omitempty"`
	Type        TypeDecl          `json:"type,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Pos         Pos               `json:"pos,omitempty"`
}

func (f Field) Position() Pos {
	return f.Pos
}

type Pos struct {
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
}

func (p Pos) ideString() string {
	return p.Filename + ":" + strconv.Itoa(p.Line)
}

type Positional interface {
	Position() Pos
}

// A TypeDef matches the specification of the language as described at https://golang.org/ref/spec#Type_definitions.
type TypeDef struct {
	Doc         string       `json:"doc,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`

	// Name of the definition
	Name string `json:"name,omitempty"`

	// ImportPath of the name
	ImportPath string `json:"importPath,omitempty"`

	// IsAlias is true plain type aliases, like e.g. MyAlias = int. The method set is retained.
	IsAlias bool `json:"isAlias,omitempty"`

	// UnderlyingType just represents the declaration as-is from the code
	UnderlyingType TypeDecl `json:"underlyingType,omitempty"`

	//	// ResolvedUnderlyingType declares which type is named. For primitives, this is always the primitive, for maps and arrays it is the type itself, see also https://stackoverflow.com/questions/29332879/golang-underlying-types#:~:text=Each%20type%20T%20has%20an,refers%20in%20its%20type%20declaration.
	//	ResolvedUnderlyingType TypeDecl `json:"resolvedUnderlyingType,omitempty"`

	Pos Pos `json:"pos,omitempty"`
}
