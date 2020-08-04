package src

type FieldBuilder struct {
	parent         *TypeBuilder
	doc            string
	name           string
	param0, param1 Qualifier
	isArray        bool
	isSlice        bool
	isChan         bool
	isMap          bool
}

func NewFieldBuilder(parent *TypeBuilder) *FieldBuilder {
	return &FieldBuilder{
		parent: parent,
	}
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

func (b *FieldBuilder) SetType(q Qualifier) *FieldBuilder {
	b.File().Use(q)
	b.param0 = q
	return b
}

func (b *FieldBuilder) SetSlice(q Qualifier) *FieldBuilder {
	b.param0 = q
	b.isSlice = true
	return b
}

func (b *FieldBuilder) Emit(w Writer) {
	t := b.File().Use(b.param0)
	if b.isSlice {
		t = "[]" + t
	}

	emitDoc(w, b.name, b.doc)
	w.Printf("%s %s\n", b.name, t)
}
