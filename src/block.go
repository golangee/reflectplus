package src

import "strings"

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

func (e *Block) Var(name string, decl *TypeDecl) *Block {
	e.AddLine("var ", name, " ", decl)
	return e
}

func (e *Block) ForEach(loopName, varName string, block *Block) *Block {
	if loopName == "" {
		loopName = guessItemName(varName)
	}
	e.AddLine("for _,", loopName, ":= range ", varName, "{")
	e.AddLine(block)
	e.AddLine("}")
	e.NewLine() // no cuddling
	return e
}

func (e *Block) If(condition string, block *Block) *Block {
	e.AddLine("if ", condition, " {")
	e.AddLine(block)
	e.AddLine("}")
	e.NewLine() // no cuddling
	return e
}

// Check is a template for
//    if <errName> != nil {
//	      return <p0,...,pn>, fmt.Errorf("<msg>: %w",<errName>)
//    }
func (e *Block) Check(errName string, msg string, returnParams ...string) *Block {
	e.AddLine("if ", errName, "!= nil {")
	e.Add("return ")
	for _, param := range returnParams {
		e.Add(param)
		e.Add(", ")
	}
	e.Add(NewTypeDecl("fmt.Errorf"), `("`, msg, `: %w", `, errName, ")")
	e.NewLine()
	e.AddLine("}")

	return e
}

func (e *Block) Add(codes ...interface{}) *Block {
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

	return e
}

func (e *Block) AddLine(codes ...interface{}) *Block {
	e.Add(codes...)

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

// guessItemName does the following:
//  receiver.Fields => field
//  Fields => field
//  Field => fieldElement
func guessItemName(name string) string {
	i := strings.Index(name, ".")
	actualName := name
	if i > -1 {
		actualName = name[i+1:]
	}

	actualName = strings.ToLower(actualName)
	if strings.HasSuffix(actualName, "s") {
		return actualName[:len(actualName)-1]
	}

	return actualName + "Element"
}
