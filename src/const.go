package src

type Const struct {
	doc      string
	name     string
	typeDecl *TypeDecl //may be nil for untyped
	value    string
	parent   FileProvider
}

func NewConst(name, rawValue string) *Const {
	return &Const{
		name:  name,
		value: rawValue,
	}
}

func NewTypedConst(name string, decl *TypeDecl, value string) *Const {
	return &Const{
		name:     name,
		typeDecl: decl,
		value:    value,
	}
}

func (t *Const) SetDoc(doc string) *Const {
	t.doc = doc
	return t
}

func (t *Const) Doc() string {
	return t.doc
}

func (t *Const) SetName(name string) *Const {
	t.name = name
	return t
}

func (t *Const) Name() string {
	return t.name
}

func (t *Const) SetType(decl *TypeDecl) *Const {
	t.typeDecl = decl
	return t
}

func (t *Const) Type() *TypeDecl {
	return t.typeDecl
}

func (t *Const) onAttach(parent FileProvider) {
	if t == nil {
		return
	}

	t.parent = parent
	if t.typeDecl != nil {
		t.typeDecl.onAttach(t)
	}
}

func (t *Const) File() *FileBuilder {
	return t.parent.File()
}

func (t *Const) Emit(w Writer) {
	emitDoc(w, t.name, t.doc)
	if t.typeDecl == nil {
		w.Printf("%s = %s\n", t.name, t.value)
	} else {
		w.Printf("%s ", t.name)
		t.typeDecl.Emit(w)
		w.Printf(" = %s\n", t.value)
	}
}
