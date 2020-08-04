package src

import "strings"



type File struct {
	imports []importStatement
}


func (f *File) emit(b strings.Builder) {
	b.WriteString("import(\n")
	for _, i := range f.imports {
		i.emit(b)
	}
	b.WriteString(")\n")
}
