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

import (
	"encoding/json"
	"strconv"
)

// Table contains all resolved type declarations and other deduplicated information. Due to the sake of the default
// value in json for integer types (==0), we use 0 to indicate "undefined".
type Table struct {
	TypeDecls map[TypeDeclId]TypeDecl
	Type      map[TypeId]Type
	Files     map[FileId]string
	Packages  map[PackageId]Package

	Types map[Qualifier]Type
}

func NewTable() *Table {
	t := &Table{
		TypeDecls: make(map[TypeDeclId]TypeDecl),
		Files:     make(map[FileId]string),
		Packages:  make(map[PackageId]Package),
		Types: map[Qualifier]Type{},
	}

	//t.initUniverse()
	return t
}

func (t *Table) HasType(q Qualifier) bool {
	_, ok := t.Types[q]
	return ok
}

func (t *Table) PutType(q Qualifier, p Type) {
	t.Types[q] = p
}

func (t *Table) initUniverse() {
	t.Packages[1] = UniversePkg
	t.TypeDecls[1] = UniverseStruct
}

func (t *Table) String() string {
	b, err := json.MarshalIndent(t, " ", " ")
	if err != nil {
		panic(err) //cannot happen
	}
	return string(b)
}

func (t *Table) TypeId(pkgId PackageId, name string) TypeDeclId {
	for k, v := range t.TypeDecls {
		if pkgId != v.PackageId {
			continue
		}

		if v.Name == name {
			return k
		}
	}

	return -1
}

// PutTypeDecl updates or inserts the declaration and returns its id.
func (t *Table) PutTypeDecl(d TypeDecl) TypeDeclId {
	newPkg, ok := t.Packages[d.PackageId]
	if !ok {
		panic("illegal state: cannot put type declaration with undeclared package" + strconv.Itoa(int(d.PackageId)))
	}

	lastId := TypeDeclId(-1)
	for k, v := range t.TypeDecls {
		if lastId < k {
			lastId = k
		}

		kPkg, ok := t.Packages[v.PackageId]
		if !ok {
			panic("illegal state: existing type declaration has undeclared package " + strconv.Itoa(int(v.PackageId)))
		}

		if newPkg.Qualifier == kPkg.Qualifier && d.Name == v.Name {
			// just update
			t.TypeDecls[k] = d
			return k
		}
	}
	lastId++

	//insert
	t.TypeDecls[lastId] = d

	return lastId
}

// PutPackageQualifier either returns an existing id or registers an empty package and returns the according new id.
func (t *Table) PutPackageQualifier(q PackageQualifier) PackageId {
	lastId := PackageId(-1)
	for k, v := range t.Packages {
		if lastId > k {
			lastId = k
		}

		if v.Qualifier == q {
			return k
		}
	}
	lastId++

	t.Packages[lastId] = Package{
		Doc:       "",
		Qualifier: q,
	}
	return lastId
}

// PutFile returns either an existing an id or puts the filename into the table and returns the according new id.
func (t *Table) PutFile(fname string) FileId {
	lastId := FileId(-1)
	for k, v := range t.Files {
		if lastId > k {
			lastId = k
		}

		if v == fname {
			return k
		}
	}
	lastId++

	t.Files[lastId] = fname
	return lastId
}
