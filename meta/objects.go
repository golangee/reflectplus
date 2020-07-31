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

package meta

import "strconv"

type ObjectId int

type TypeName struct {
	Pos
	PackageId
	Doc         string
	Annotations []Annotation
	Underlying  TypeId
	Name        string
}

// A TypeDeclId refers to a concrete declaration in source and refers always to (invariant) underlying types.
type TypeDeclId int

// A TypeDecl is a type declaration and binds an identifier, the type name, to a type.
// Type declarations are either an alias declaration or a type definition.
type TypeDecl struct {
	Pos
	PackageId
	Doc         string
	Annotations []Annotation
	Name        string
	Underlying  TypeId
	// Methods contains all concrete declared methods TODO what is with methods from embedded types?
	Methods []DeclMethod

	// Fields are only expressible by Structs
	Fields []DeclField
}

type DeclField struct {
	Pos
	Doc  string
	Name string
	Type TypeDeclId
	Tags map[string]string
}

type DeclMethod struct {
	Pos
	Doc          string
	Annotations  []Annotation
	ReceiverName string
	Receiver     TypeDeclId
	Name         string
	Params       []Param
	Results      []Param
}

type DeclParam struct {
	Pos
	Doc         string
	Annotations []Annotation
	Name        string
}

type DeclInterface struct {
	Pos
	Name string
}

type FileId int

type Pos struct {
	FileId FileId
	Line   int
	Column int
}

type Location string

func NewLocation(filename string, line, col int) Location {
	return Location(filename + ":" + strconv.Itoa(line) + ":" + strconv.Itoa(col))
}

type PackageId int

// A PackageQualifier consists of an import path and the according package name. This is rather obscure, because
// you can never deduce the actual package name from its path, however it allows at least elegant solutions like this:
//  * versioned packages, e.g. github.com/myproject/myapi and github.com/myproject/myapi/v2 should both be named myapi
//  * using path with chars which are illegal identifiers, like github.com/myproject/my-api (could still be myapi)
type PackageQualifier struct {
	// Path is a / separated name to resolve a package.
	Path string
	// Name is the actual identifier of the package. This is also used by default, if not renamed at import.
	Name string
}

type Package struct {
	Doc       string
	Qualifier PackageQualifier
}

type Annotation struct {
	Pos Location
	Doc    string
	Name   string
	Values map[string]interface{}
}
