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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golangee/reflectplus/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Generate(dir string) error {
	if !strings.HasPrefix(dir, string(os.PathSeparator)) {
		absDir, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = filepath.Join(absDir, dir)
	}

	fmt.Printf("scanning %s\n", dir)

	pkg, err := parser.ParsePackage(nil, dir)
	if err != nil {
		return err
	}

	metaPkg, err := ParseMetaModel(pkg)
	if err != nil {
		return err
	}
	//fmt.Println(metaPkg)

	return writeReflectFile(dir, metaPkg.ImportPath, *metaPkg)
}

func writeReflectFile(dir string, importPath string, pkg Package) error {
	g := newGoGenFile(importPath)
	writeInit(g, pkg)
	for _, iface := range pkg.AllInterfaces() {
		generateSrcProxy(g, iface)
	}
	writeReflectionData(g, pkg)
	return ioutil.WriteFile(filepath.Join(dir, "reflect.gen.go"), []byte(g.FormatString()), os.ModePerm)
}

func writeInit(w *goGenFile, pkg Package) {
	w.Printf("func init(){\n")

	// write the proxy registrations
	for _, iface := range pkg.AllInterfaces() {
		tName := typesafeName(iface.ImportPath) + iface.Name + "Proxy"
		w.Printf(`%s("%s","%s", func(h %s) interface{} {`, w.ImportName(importPathReflectPlus, "AddProxyFactory"), iface.ImportPath, iface.Name, w.ImportName(importPathReflectPlus, "InvocationHandler"))
		w.Printf("\n")
		w.Printf("return %s{Handler:h}\n", tName)
		w.Printf("})\n")
	}

	// write the package information
	w.Printf("if _,err := %s(metaData);err!=nil{\n", w.ImportName(importPathReflectPlus, "ImportMetaData"))
	w.Printf("panic(err)\n")
	w.Printf("}\n")

	// write golang reflect connector
	writeTypeOfRegistrations(w, pkg)
	w.Printf("}\n")
}

func writeReflectionData(w *goGenFile, pkg Package) {
	dat, err := json.Marshal(&pkg)
	if err != nil {
		panic(err) // cannot happen
	}
	// this way we avoid ugly escaping, however we should emit structs directly instead
	w.Printf("var metaData,_ = %s.DecodeString(\"%s\")", w.ImportName("encoding/base64", "StdEncoding"), base64.StdEncoding.EncodeToString(dat))
}

func writeTypeOfRegistrations(w *goGenFile, pkg Package) {
	for _, s := range pkg.AllStructs() {
		if strings.HasSuffix(s.ImportPath, "main") { //TODO this is wrong any way, importpath is the dir name, not the actual parsed package name
			continue //TODO  avoid import "xyz" is a program, not an importable package
		}
		typeName := w.ImportName(s.ImportPath, s.Name)
		add := w.ImportName(importPathReflectPlus, "AddType")
		refl := w.ImportName("reflect", "TypeOf")
		w.Printf("%s(\"%s\",\"%s\",%s(%s{}))\n", add, s.ImportPath, s.Name, refl, typeName)
	}
}
