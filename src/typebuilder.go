package src

import (
	"github.com/golangee/reflectplus/meta"
)

type internalType int

const (
	typeBase internalType = iota
	typeStruct
	typeInterface
)

type TypeBuilder struct {
	parent  *FileBuilder
	doc     string
	name    string
	rhs     *meta.Type
	iType   internalType
	methods []*FuncBuilder
	fields  []*FieldBuilder
}

func NewType() *TypeBuilder {
	return &TypeBuilder{}
}

func NewStruct(name string) *TypeBuilder {
	return &TypeBuilder{
		name:  name,
		iType: typeStruct,
	}
}

func NewInterface(name string) *TypeBuilder {
	return &TypeBuilder{
		name:  name,
		iType: typeInterface,
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

func (b *TypeBuilder) Doc() string {
	return b.doc
}

func (b *TypeBuilder) Name() string {
	return b.name
}

func (b *TypeBuilder) Methods() []*FuncBuilder {
	return b.methods
}

func (b *TypeBuilder) Fields() []*FieldBuilder {
	return b.fields
}

func (b *TypeBuilder) AddFields(fields ...*FieldBuilder) *TypeBuilder {
	b.iType = typeStruct
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
	switch b.iType {
	case typeStruct:
		w.Printf(" struct {\n")
		for _, field := range b.fields {
			field.Emit(w)
		}
		w.Printf("}\n")
	case typeBase:
		w.Printf("\n")
	case typeInterface:
		w.Printf(" interface {\n")

		for _, method := range b.methods {
			method.Emit(w)
			w.Printf("\n")
		}

		w.Printf("}\n")
	}

	if b.iType != typeInterface {
		for _, method := range b.methods {
			method.Emit(w)
			w.Printf("\n")
		}
	}

}
