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
	"reflect"
	"strings"
)

var packages []Package
var typesByName map[string]reflect.Type = make(map[string]reflect.Type)

func AddPackage(pkg Package) {
	packages = append(packages, pkg)
}

func ImportMetaData(jsn []byte) (Package, error) {
	pkg := Package{}
	err := json.Unmarshal(jsn, &pkg)
	if err != nil {
		return pkg, err
	}
	AddPackage(pkg)
	return pkg, nil
}

func Packages() []Package {
	return packages
}

// FindType returns the type or nil
func FindType(importPath string, name string) reflect.Type {
	return typesByName[importPath+"#"+name]
}

func AddType(importPath string, name string, p reflect.Type) {
	typesByName[importPath+"#"+name] = p
}

// FindByType tries to find the Struct or interface from the reflect type, otherwise returns nil.
func FindByType(t reflect.Type) interface{} {
	for k, v := range typesByName {
		if v == t {
			tokens := strings.Split(k, "#")
			strct := FindStruct(tokens[0], tokens[1])
			if strct != nil {
				return strct
			}
			iface := FindInterface(tokens[0], tokens[1])
			if iface != nil {
				return iface
			}
			return nil
		}
	}
	return nil
}

func Interfaces() []Interface {
	var res []Interface
	for _, p := range packages {
		for _, iface := range p.AllInterfaces() {
			res = append(res, iface)
		}
	}
	return res
}

func FindInterface(importPath string, name string) *Interface {
	for _, p := range packages {
		for _, iface := range p.AllInterfaces() {
			if iface.ImportPath == importPath && iface.Name == name {
				return &iface
			}
		}
	}
	return nil
}

func FindStruct(importPath string, name string) *Struct {
	for _, p := range packages {
		for _, iface := range p.AllStructs() {
			if iface.ImportPath == importPath && iface.Name == name {
				return &iface
			}
		}
	}
	return nil
}

func FindPackage(importPath string) *Package {
	for _, p := range packages {
		if p.ImportPath == importPath {
			return &p
		}
		var r *Package
		p.VisitPackages(func(pkg Package) bool {
			if pkg.ImportPath == importPath {
				r = &pkg
				return false
			}
			return true
		})
		if r != nil {
			return r
		}
	}
	return nil
}
