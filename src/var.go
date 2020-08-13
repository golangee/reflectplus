package src

type Var struct {
	doc      string
	name     string
	rhs      *Block
	typeDecl *TypeDecl
	parent   FileProvider
}

func NewVar(name string) *Var {
	return &Var{
		name: name,
	}
}

func (t *Var) SetRHS(block *Block) *Var {
	t.rhs = block
	t.rhs.onAttach(t)
	return t
}

func (t *Var) RHS() *Block {
	return t.rhs
}

func (t *Var) SetDoc(doc string) *Var {
	t.doc = doc
	return t
}

func (t *Var) Doc() string {
	return t.doc
}

func (t *Var) SetName(name string) *Var {
	t.name = name
	return t
}

func (t *Var) Name() string {
	return t.name
}

func (t *Var) SetType(decl *TypeDecl) *Var {
	t.typeDecl = decl
	return t
}

func (t *Var) Type() *TypeDecl {
	return t.typeDecl
}

func (t *Var) onAttach(parent FileProvider) {
	if t == nil {
		return
	}

	t.parent = parent
	if t.typeDecl != nil {
		t.typeDecl.onAttach(parent)
	}
}

func (t *Var) File() *FileBuilder {
	return t.parent.File()
}

func (t *Var) Emit(w Writer) {
	emitDoc(w, t.name, t.doc)
	w.Printf(t.name)
	w.Printf(" ")
	if t.typeDecl != nil {
		t.typeDecl.Emit(w)
		w.Printf(" ")
	}
	w.Printf("= ")
	t.rhs.Emit(w)
}
