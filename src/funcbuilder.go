package src

type FuncBuilder struct {
	parent  *TypeBuilder
	doc     string
	name    string
	recName string
}

func NewFuncBuilder(parent *TypeBuilder) *FuncBuilder {
	return &FuncBuilder{parent: parent}
}

func (b *FuncBuilder) SetName(name string) *FuncBuilder {
	b.name = name
	return b
}

func (b *FuncBuilder) SetDoc(doc string) *FuncBuilder {
	b.doc = doc
	return b
}

func (b *FuncBuilder) Emit(w Writer) {
	emitDoc(w, b.name, b.doc)
	w.Printf("func ")
	if b.parent != nil {
		if b.recName == "" {
			b.recName = "_self" //TODO
		}
		w.Printf("(%s %s) ", b.recName, b.name)
	}

	w.Printf("%s(", b.name)
	w.Printf(")")
	w.Printf("{\n")
	w.Printf("}\n")
}
