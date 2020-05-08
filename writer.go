package reflectplus

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"
)

type newImportName string
type origImportPath string

type goGenFile struct {
	importPath   string
	namedImports map[origImportPath]newImportName
	sb           *strings.Builder
	indent       int
	newLine      bool
}

func newGoGenFile(importPath string) *goGenFile {
	return &goGenFile{sb: &strings.Builder{}, importPath: importPath, namedImports: make(map[origImportPath]newImportName)}
}

func (w *goGenFile) Import(importPath string) string {
	if importPath == w.importPath || importPath == "" {
		return ""
	}
	newImport, has := w.namedImports[origImportPath(importPath)]
	if has {
		return string(newImport)
	}

	for i := 1; i > 0; i++ {
		newName := lastName(importPath)
		if i > 1 {
			newName += strconv.Itoa(i)
		}
		if !w.HasImport(newName) {
			w.namedImports[origImportPath(importPath)] = newImportName(newName)
			return newName
		}
	}
	panic("unreachable")
}

func (w *goGenFile) ImportName(importPath string, name string) string {
	renamedImport := w.Import(importPath)
	if renamedImport == "" {
		return name
	}
	return renamedImport + "." + name
}

func (w *goGenFile) HasImport(importPath string) bool {
	_, has := w.namedImports[origImportPath(importPath)]
	return has
}

func (w *goGenFile) Indent(i int) {
	w.indent += i
}

func (w *goGenFile) ShiftLeft() {
	w.Indent(-2)
}

func (w *goGenFile) ShiftRight() {
	w.Indent(2)
}

func (w *goGenFile) Printf(str string, args ...interface{}) {
	if w.newLine {
		for i := 0; i < w.indent; i++ {
			w.sb.WriteByte(' ')
		}
	}
	w.sb.WriteString(fmt.Sprintf(str, args...))
	w.newLine = strings.HasSuffix(str, "\n")
}

func (w *goGenFile) String() string {
	pkgname := lastName(w.importPath)
	tmp := &strings.Builder{}
	tmp.WriteString("// Code generated by reflectplus. DO NOT EDIT.\n\n")
	tmp.WriteString("package " + pkgname + "\n\n")
	tmp.WriteString("import (\n")
	for importPath, importName := range w.namedImports {
		tmp.WriteString(string(importName) + " \"" + string(importPath) + "\"\n")
	}
	tmp.WriteString(")\n")
	tmp.WriteString(w.sb.String())
	return tmp.String()
}

func (w *goGenFile) FormatString() string {
	b, err := format.Source([]byte(w.String()))
	if err != nil {
		fmt.Println(w.String())
		panic(err)
	}
	return string(b)
}

func lastName(text string) string {
	tokens := strings.Split(text, "/")
	pkgname := text
	if len(tokens) > 0 {
		pkgname = tokens[len(tokens)-1]
	}
	return pkgname
}
