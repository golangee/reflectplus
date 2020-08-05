package src

type Emitter interface {
	Emit(w Writer)
}

type fileProviderAttacher interface{
	onAttach(parent FileProvider)
	Emit(w Writer)
}

type SPrintf struct {
	Str  string
	Args []interface{}
}

func (s SPrintf) Emit(w Writer) {
	w.Printf(s.Str, s.Args...)
}
