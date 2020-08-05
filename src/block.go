package src

type Block struct {
	parent  FileProvider
	emitter []Emitter
}

func NewBlock(line ...interface{}) *Block {
	e := &Block{}
	e.AddLine(line...)
	return e
}

func (e *Block) File() *FileBuilder {
	return e.parent.File()
}

func (e *Block) NewLine() *Block {
	e.emitter = append(e.emitter, SPrintf{
		Str: "\n",
	})
	return e
}

func (e *Block) AddLine(codes ...interface{}) *Block {
	for _, code := range codes {
		switch t := code.(type) {
		case fileProviderAttacher:
			e.emitter = append(e.emitter, t)
			t.onAttach(e)
		case Emitter:
			e.emitter = append(e.emitter, t)
		default:
			e.emitter = append(e.emitter, SPrintf{
				Str:  "%v",
				Args: []interface{}{t},
			})
		}
	}
	if len(codes) > 0 {
		e.NewLine()
	}

	return e
}

func (e *Block) onAttach(parent FileProvider) {
	e.parent = parent
}

func (e *Block) Emit(w Writer) {
	for _, v := range e.emitter {
		v.Emit(w)
	}
}
