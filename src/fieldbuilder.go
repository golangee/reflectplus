package src

type FieldBuilder struct {
	parent *TypeBuilder
	doc    string
	name   string
	decl   *TypeDecl
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

func (b *FieldBuilder) File() *FileBuilder {
	return b.parent.File()
}

func (b *FieldBuilder) SetType(t *TypeDecl) *FieldBuilder {
	b.decl = t
	return b
}

func (b *FieldBuilder) Emit(w Writer) {
	emitDoc(w, b.name, b.doc)
	w.Printf(b.name)
	w.Printf(" ")
	b.decl.Emit(w)
	w.Printf("\n")
}
