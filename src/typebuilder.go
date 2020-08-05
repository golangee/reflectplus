package src

import (
	"github.com/golangee/reflectplus/meta"
)

type TypeBuilder struct {
	parent   *FileBuilder
	doc      string
	name     string
	rhs      *meta.Type
	isStruct bool
	methods  []*FuncBuilder
	fields   []*FieldBuilder
}

func NewType() *TypeBuilder {
	return &TypeBuilder{}
}

func NewStruct(name string) *TypeBuilder {
	return &TypeBuilder{
		name:     name,
		isStruct: true,
	}
}

func (b *TypeBuilder) onAttach(parent *FileBuilder) {
	b.parent = parent
}

func (b *TypeBuilder) SetName(name string) *TypeBuilder {
	b.name = name
	return b
}

func (b *TypeBuilder) SetDoc(doc string) *TypeBuilder {
	b.doc = doc
	return b
}

func (b *TypeBuilder) AddFields(fields ...*FieldBuilder) *TypeBuilder {
	b.isStruct = true
	b.fields = append(b.fields, fields...)
	for _, f := range fields {
		f.onAttach(b)
	}
	return b
}

func (b *TypeBuilder) AddMethods(methods ...*FuncBuilder) *TypeBuilder {
	b.methods = append(b.methods, methods...)
	for _, m := range methods {
		m.onAttach(b)
	}
	return b
}

func (b *TypeBuilder) File() *FileBuilder {
	return b.parent
}

func (b *TypeBuilder) Emit(w Writer) {
	emitDoc(w, b.name, b.doc)
	w.Printf("type %s", b.name)
	if b.isStruct {
		w.Printf(" struct {\n")
		for _, field := range b.fields {
			field.Emit(w)
		}
		w.Printf("}\n")
	} else {
		w.Printf("\n")
	}

	for _, method := range b.methods {
		method.Emit(w)
	}
}
