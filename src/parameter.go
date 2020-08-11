package src

type Parameter struct {
	name   string
	decl   *TypeDecl
	parent FileProvider
}

func NewParameter(name string, decl *TypeDecl) *Parameter {
	return &Parameter{
		name: name,
		decl: decl,
	}
}

func (p *Parameter) Name() string {
	return p.name
}

func (p *Parameter) Decl() *TypeDecl {
	return p.decl
}

func (p *Parameter) onAttach(parent FileProvider) {
	p.parent = parent
	p.decl.onAttach(parent)
}

func (p *Parameter) Emit(w Writer) {
	w.Printf(p.name)
	w.Printf(" ")
	p.decl.Emit(w)
}

func (p *Parameter) emitAsVariadic(w Writer) {
	w.Printf(p.name)
	w.Printf("...")
	p.decl.Emit(w)
}
