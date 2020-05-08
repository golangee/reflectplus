package reflectplus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/worldiety/reflectplus/parser"
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
		w.Printf(`%s("%s","%s", func(h %s) interface{} {`, w.ImportName("github.com/worldiety/reflectplus", "AddProxyFactory"), iface.ImportPath, iface.Name, w.ImportName("github.com/worldiety/reflectplus", "InvocationHandler"))
		w.Printf("\n")
		w.Printf("return %s{Handler:h}\n", tName)
		w.Printf("})\n")
	}

	// write the package information
	w.Printf("if _,err := %s(metaData);err!=nil{\n", w.ImportName("github.com/worldiety/reflectplus", "ImportMetaData"))
	w.Printf("panic(err)\n")
	w.Printf("}\n")

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
