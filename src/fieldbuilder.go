package src

import (
	"strconv"
	"strings"
)

type FieldBuilder struct {
	parent *TypeBuilder
	doc    string
	name   string
	decl   *TypeDecl
	tags   map[string][]string
}

func NewField(name string, typeDecl *TypeDecl) *FieldBuilder {
	b := &FieldBuilder{
		name: name,
		decl: typeDecl,
	}
	return b
}

func (b *FieldBuilder) onAttach(parent *TypeBuilder) {
	b.parent = parent
	b.decl.onAttach(parent)
}

func (b *FieldBuilder) SetDoc(doc string) *FieldBuilder {
	b.doc = doc
	return b
}

func (b *FieldBuilder) SetName(name string) *FieldBuilder {
	b.name = name
	return b
}

func (b *FieldBuilder) Name() string {
	return b.name
}

func (b *FieldBuilder) Type() *TypeDecl {
	return b.decl
}

func (b *FieldBuilder) Doc() string {
	return b.doc
}

func (b *FieldBuilder) Tags() map[string][]string {
	return b.tags
}

func (b *FieldBuilder) File() *FileBuilder {
	return b.parent.File()
}

func (b *FieldBuilder) SetType(t *TypeDecl) *FieldBuilder {
	b.decl = t
	return b
}

func (b *FieldBuilder) AddTag(key string, values ...string) *FieldBuilder {
	if b.tags == nil {
		b.tags = map[string][]string{}
	}

	vals := b.tags[key]
	vals = append(vals, values...)
	b.tags[key] = vals

	return b
}

func (b *FieldBuilder) Emit(w Writer) {
	emitDoc(w, b.name, b.doc)
	w.Printf(b.name)
	w.Printf(" ")
	b.decl.Emit(w)
	if len(b.tags) > 0 {
		sb := &strings.Builder{}
		sb.WriteRune('`')
		for k, v := range b.tags {
			sb.WriteString(k)
			sb.WriteRune(':')
			values := strings.Join(v, ",")
			sb.WriteString(strconv.Quote(values))
			sb.WriteRune(' ')
		}
		sb.WriteRune('`')
		w.Printf(sb.String())
	}
	w.Printf("\n")
}
