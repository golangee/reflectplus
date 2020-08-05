package src

// Types and generics are expressed by a type declaration. For example:
//   int: TypeDecl{qualifier:"int"}
//     is equal to NewTypeDecl("int")
//   []int: TypeDecl{qualifier:"[]",params:[]TypeDecl{qualifier:"int"}}
//     is equal to NewSliceDecl("int")
type TypeDecl struct {
	qualifier Qualifier
	params    []*TypeDecl
	parent    FileProvider
}

func NewTypeDecl(qualifier Qualifier) *TypeDecl {
	return &TypeDecl{qualifier: qualifier}
}

func NewGenericDecl(qualifier Qualifier, params ...*TypeDecl) *TypeDecl {
	return &TypeDecl{
		qualifier: qualifier,
		params:    params,
	}
}

func NewSliceDecl(t *TypeDecl) *TypeDecl {
	return NewGenericDecl("[]", t)
}

func NewMapDecl(key, val *TypeDecl) *TypeDecl {
	return NewGenericDecl("map", key, val)
}

func NewPointerDecl(t *TypeDecl) *TypeDecl {
	return NewGenericDecl("*", t)
}

func NewErrorDecl() *TypeDecl {
	return NewTypeDecl("error")
}

func (t *TypeDecl) onAttach(parent FileProvider) {
	if t == nil {
		return
	}

	t.parent = parent
	for _, p := range t.params {
		p.onAttach(parent)
	}
}

func (t TypeDecl) Emit(w Writer) {
	base := t.parent.File().Use(t.qualifier)
	w.Printf(base)
	w.Printf(" ")
	switch base {
	case "map":
		w.Printf("[")
		t.params[0].Emit(w)
		w.Printf("]")
		t.params[1].Emit(w)
	default:
		for _, p := range t.params {
			p.Emit(w)
			w.Printf(" ")
		}
	}

}
