package src

import "strings"

type FuncBuilder struct {
	parent        *TypeBuilder
	doc           string
	name          string
	recName       string
	isPtrReceiver bool
	params        []*Parameter
	results       []*Parameter
	body          []*Block
}

func NewFunc(name string) *FuncBuilder {
	b := &FuncBuilder{}
	b.SetName(name)
	return b
}

func (b *FuncBuilder) Name() string {
	return b.name
}

func (b *FuncBuilder) Params() []*Parameter {
	return b.params
}

func (b *FuncBuilder) AddParams(params ...*Parameter) *FuncBuilder {
	b.params = append(b.params, params...)
	for _, p := range params {
		p.onAttach(b)
	}

	return b
}

func (b *FuncBuilder) AddResults(params ...*Parameter) *FuncBuilder {
	b.results = append(b.results, params...)
	for _, p := range params {
		p.onAttach(b)
	}

	return b
}

func (b *FuncBuilder) Results() []*Parameter {
	return b.results
}

func (b *FuncBuilder) AddBody(blocks ...*Block) *FuncBuilder {
	b.body = append(b.body, blocks...)
	for _, block := range blocks {
		block.onAttach(b)
	}
	return b
}

func (b *FuncBuilder) File() *FileBuilder {
	return b.parent.File()
}

func (b *FuncBuilder) onAttach(parent *TypeBuilder) {
	b.parent = parent

	if b.recName == "" && len(parent.name) > 0 {
		b.recName = strings.ToLower(parent.name[0:1])
	}
}

func (b *FuncBuilder) SetName(name string) *FuncBuilder {
	b.name = name
	return b
}

func (b *FuncBuilder) SetReceiverName(name string) *FuncBuilder {
	b.recName = name
	return b
}

func (b *FuncBuilder) SetPointerReceiver(isPtrRec bool) *FuncBuilder {
	b.isPtrReceiver = isPtrRec
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
		ptrRec := ""
		if b.isPtrReceiver {
			ptrRec = "*"
		}
		w.Printf("(%s %s%s) ", b.recName, ptrRec, b.parent.name)
	}

	w.Printf("%s(", b.name)
	for _, p := range b.params {
		p.Emit(w)
		w.Printf(",")
	}
	w.Printf(")")

	w.Printf("(")
	for _, p := range b.results {
		p.Emit(w)
		w.Printf(",")
	}
	w.Printf(")")

	w.Printf("{\n")
	for _, block := range b.body {
		block.Emit(w)
	}
	w.Printf("}\n")
}
