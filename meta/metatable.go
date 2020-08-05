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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"sort"
	"strconv"
)

// DeclId is usually a unique hash for a declaration (not just named types).
type DeclId string

type DeclIdBuilder struct {
	hasher hash.Hash
}

func NewDeclId() *DeclIdBuilder {
	return &DeclIdBuilder{hasher: sha256.New()}
}

func (b *DeclIdBuilder) Put(values ...interface{}) *DeclIdBuilder {
	for _, v := range values {
		switch t := v.(type) {
		case string:
			b.hasher.Write([]byte(t))
		case int:
			b.hasher.Write([]byte(strconv.Itoa(t)))
		default:
			b.hasher.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}

	return b
}

func (b *DeclIdBuilder) Finish() DeclId {
	r := b.hasher.Sum(nil)
	b.hasher.Reset()
	return DeclId(hex.EncodeToString(r[:16])) // 16 byte/128bit is probably still more than enough
}

type PkgId string

type Package struct {
	Path         string
	Name         string
	Declarations []DeclId
}

// Table contains all resolved type declarations and other deduplicated information. Due to the sake of the default
// value in json for integer types (==0), we use 0 to indicate "undefined".
type Table struct {
	Packages     map[PkgId]*Package
	Declarations map[DeclId]Type
}

func NewTable() *Table {
	t := &Table{
		Packages:     map[PkgId]*Package{},
		Declarations: map[DeclId]Type{},
	}

	return t
}

// DeclIds returns a stable and sorted slice of declarations ids
func (t *Table) DeclIds() []DeclId {
	tmp := make([]DeclId, 0, len(t.Declarations))
	for k := range t.Declarations {
		tmp = append(tmp, k)
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i] < tmp[j]
	})

	return tmp
}

func (t *Table) HasDeclaration(q DeclId) bool {
	_, ok := t.Declarations[q]
	return ok
}

func (t *Table) PutDeclaration(q DeclId, p Type) {
	t.Declarations[q] = p
}

func (t *Table) PutNamedDeclaration(importPath, pkgName string, q DeclId, p *Named) {
	pid, ok := t.PackageByImportPath(importPath)
	if !ok {
		pid = PkgId(NewDeclId().Put(importPath).Finish())
		t.Packages[pid] = &Package{
			Path: importPath,
			Name: pkgName,
		}
	}
	pkg := t.Packages[pid]

	if pkg.Name != pkgName {
		panic("inconsistent package name:" + pkgName + " vs " + pkg.Name)
	}

	t.PutDeclaration(q, Type{
		Named: p,
	})
	pkg.Declarations = append(pkg.Declarations, q)
}

// CreateImportTable creates a new table which assigns each declaration id to its containing package id.
func (t *Table) CreateImportTable() map[DeclId]PkgId {
	r := map[DeclId]PkgId{}
	for pid, pkg := range t.Packages {
		for _, did := range pkg.Declarations {
			r[did] = pid
		}
	}
	return r
}

func (t *Table) PackageByImportPath(importPath string) (PkgId, bool) {
	for k, v := range t.Packages {
		if v.Path == importPath {
			return k, true
		}
	}

	return "", false
}

func (t *Table) String() string {
	b, err := json.MarshalIndent(t, " ", " ")
	if err != nil {
		panic(err) //cannot happen
	}
	return string(b)
}
