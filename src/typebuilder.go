package src

import "strconv"

type internalType int

const (
	typeBase internalType = iota
	typeStruct
	typeInterface
)

type TypeBuilder struct {
	parent     *FileBuilder
	doc        string
	name       string
	iType      internalType
	methods    []*FuncBuilder
	fields     []*FieldBuilder
	baseType   *TypeDecl
	enumValues []string
	consts     []*Const //associated constants
	vars       []*Var   //associated variables
}

func NewType() *TypeBuilder {
	return &TypeBuilder{}
}

func NewStruct(name string) *TypeBuilder {
	return &TypeBuilder{
		name:  name,
		iType: typeStruct,
	}
}

func NewInterface(name string) *TypeBuilder {
	return &TypeBuilder{
		name:  name,
		iType: typeInterface,
	}
}

func NewBaseType(name string, from *TypeDecl) *TypeBuilder {
	b := &TypeBuilder{
		name:     name,
		iType:    typeBase,
		baseType: from,
	}

	b.baseType.onAttach(b)
	return b
}

// NewIntEnum generates a go-idiomatic and efficient based enum-like type with int16 as underlying type.
func NewIntEnum(name string, values ...string) *TypeBuilder {
	switchCase := NewBlock()
	switchCase.AddLine("switch e {")
	for i, value := range values {
		enumNum := i + 1
		switchCase.AddLine("case ", enumNum, ":")
		switchCase.AddLine("return ", strconv.Quote(value))
	}
	switchCase.AddLine("default:")
	switchCase.AddLine("return ", NewTypeDecl("strconv.Itoa(int(e))"))
	switchCase.AddLine("}")

	b := NewBaseType(name, NewTypeDecl("int16"))
	b.AddMethods(
		NewFunc("String").
			SetDoc("...returns the enums name.").
			SetReceiverName("e").
			AddResults(NewParameter("", NewTypeDecl("string"))).
			AddBody(switchCase),
		NewFunc("IsValid").
			SetDoc("...returns true, if the value represents a defined enum value.").
			SetReceiverName("e").
			AddResults(NewParameter("", NewTypeDecl("bool"))).
			AddBody(NewBlock("return e >= ", 1, " && e <= ", len(values))),
	)

	for i, value := range values {
		b.AddConstants(NewTypedConst(name+snakeCaseToCamelCase(value), NewTypeDecl(Qualifier(name)), strconv.Itoa(i+1)).
			SetDoc("...represents the enum '" + value + "'."),
		)
	}

	arrayVarExpr := NewBlock().Add("[]", name, "{")
	for i := range values {
		arrayVarExpr.Add(i, ",")
	}
	arrayVarExpr.Add("}")
	b.AddVariables(NewVar(name + "Values").SetDoc("...contains all valid enum states.").SetRHS(arrayVarExpr))

	return b
}

// NewStringEnum generates a go-idiomatic enum-like type with string as underlying type.
func NewStringEnum(name string, values ...string) *TypeBuilder {
	switchCase := NewBlock()
	switchCase.AddLine("switch e {")
	for _, value := range values {
		switchCase.AddLine("case ", strconv.Quote(value), ":")
		switchCase.AddLine("return true")
	}
	switchCase.AddLine("default:")
	switchCase.AddLine("return false")
	switchCase.AddLine("}")

	b := NewBaseType(name, NewTypeDecl("string"))
	b.AddMethods(
		NewFunc("String").
			SetDoc("...returns the enums name.").
			SetReceiverName("e").
			AddResults(NewParameter("", NewTypeDecl("string"))).
			AddBody(NewBlock("return string(e)")),
		NewFunc("IsValid").
			SetDoc("...returns true, if the value represents a defined enum value.").
			SetReceiverName("e").
			AddResults(NewParameter("", NewTypeDecl("bool"))).
			AddBody(switchCase),
	)

	for _, value := range values {
		b.AddConstants(NewTypedConst(name+snakeCaseToCamelCase(value), NewTypeDecl(Qualifier(name)), strconv.Quote(value)).
			SetDoc("...represents the enum '" + value + "'."),
		)
	}

	arrayVarExpr := NewBlock().Add("[]", name, "{")
	for _, value := range values {
		arrayVarExpr.Add(strconv.Quote(value), ",")
	}
	arrayVarExpr.Add("}")
	b.AddVariables(NewVar(name + "Values").SetDoc("...contains all valid enum states.").SetRHS(arrayVarExpr))

	return b
}

func (b *TypeBuilder) Constants() []*Const {
	return b.consts
}

// AddConst adds associated constant declarations
func (b *TypeBuilder) AddConstants(constExpr ...*Const) *TypeBuilder {
	b.consts = append(b.consts, constExpr...)
	for _, c := range constExpr {
		c.onAttach(b)
	}

	return b
}

func (b *TypeBuilder) Variables() []*Var {
	return b.vars
}

// AddVariables adds associated variable declarations
func (b *TypeBuilder) AddVariables(varExpr ...*Var) *TypeBuilder {
	b.vars = append(b.vars, varExpr...)
	for _, c := range varExpr {
		c.onAttach(b)
	}

	return b
}

func (b *TypeBuilder) onAttach(parent *FileBuilder) {
	b.parent = parent
}

func (b *TypeBuilder) SetName(name string) *TypeBuilder {
	b.name = name
	return b
}

func (b *TypeBuilder) SetDoc(doc string) *TypeBuilder {
	b.doc = doc
	return b
}

func (b *TypeBuilder) Doc() string {
	return b.doc
}

func (b *TypeBuilder) Name() string {
	return b.name
}

func (b *TypeBuilder) Methods() []*FuncBuilder {
	return b.methods
}

func (b *TypeBuilder) Fields() []*FieldBuilder {
	return b.fields
}

func (b *TypeBuilder) AddFields(fields ...*FieldBuilder) *TypeBuilder {
	b.iType = typeStruct
	b.fields = append(b.fields, fields...)
	for _, f := range fields {
		f.onAttach(b)
	}
	return b
}

func (b *TypeBuilder) AddMethods(methods ...*FuncBuilder) *TypeBuilder {
	b.methods = append(b.methods, methods...)
	for _, m := range methods {
		m.onAttach(b)
	}
	return b
}

func (b *TypeBuilder) File() *FileBuilder {
	return b.parent
}

func (b *TypeBuilder) Emit(w Writer) {
	emitDoc(w, b.name, b.doc)
	w.Printf("type %s", b.name)
	switch b.iType {
	case typeStruct:
		w.Printf(" struct {\n")
		for _, field := range b.fields {
			field.Emit(w)
		}
		w.Printf("}\n")
	case typeBase:
		w.Printf(" ")
		b.baseType.Emit(w)
		w.Printf("\n")
	case typeInterface:
		w.Printf(" interface {\n")

		for _, method := range b.methods {
			method.Emit(w)
			w.Printf("\n")
		}

		w.Printf("}\n")
	}

	if b.iType != typeInterface {
		for _, method := range b.methods {
			method.Emit(w)
			w.Printf("\n")
		}
	}

	w.Printf("\n")

	if len(b.consts) > 0 {
		w.Printf("const(\n")
		for _, c := range b.consts {
			c.Emit(w)
		}
		w.Printf(")\n")
		w.Printf("\n")
	}

	if len(b.vars) > 0 {
		w.Printf("var(\n")
		for _, c := range b.vars {
			c.Emit(w)
		}
		w.Printf(")\n")
		w.Printf("\n")
	}

}
