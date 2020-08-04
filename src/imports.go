package src

import "strings"

type packageQualifier struct {
	Path string
	Name string
}

type importStatement struct {
	pkg  packageQualifier
	name string
}

func (stmt *importStatement) emit(b strings.Builder) {
	if stmt.name == "" {
		b.WriteString(stmt.pkg.Name)
	} else {
		b.WriteString(stmt.name)
	}

	b.WriteString(" \"")
	b.WriteString(stmt.pkg.Path)
	b.WriteString("\"\n")
}
