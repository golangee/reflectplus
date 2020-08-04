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

func NewTypeBuilder(parent *FileBuilder) *TypeBuilder {
	return &TypeBuilder{
		parent: parent,
	}
}

func (b *TypeBuilder) SetName(name string) *TypeBuilder {
	b.name = name
	return b
}

func (b *TypeBuilder) SetDoc(doc string) *TypeBuilder {
	b.doc = doc
	return b
}

func (b *TypeBuilder) AddField(f func(f *FieldBuilder)) *TypeBuilder {
	b.isStruct = true
	builder := NewFieldBuilder(b)
	b.fields = append(b.fields, builder)
	f(builder)
	return b
}

func (b *TypeBuilder) AddMethod(f func(f *FuncBuilder)) *TypeBuilder {
	fb := NewFuncBuilder(b)
	f(fb)
	b.methods = append(b.methods, fb)
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

	for _,method := range b.methods{
		method.Emit(w)
	}
}
