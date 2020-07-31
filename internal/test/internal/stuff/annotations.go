package stuff

// An AnnotatedStruct carries annotations like
// @Test("hello")
// @ee.Test("hello":"world", "hello":"dude")
type AnnotatedStruct struct {
	// Another field comment
	// @FieldAnnotation("hello field")
	SomeField string `json:"name,omitempty"`
}

// An AnnotatedStruct2 carries other annotations like
// @Test2("hello2")
type AnnotatedStruct2 struct {
	// Shares the same underlying type as AnnotatedStruct but the first level must be different
	// @FieldAnnotation2("hello field2")
	SomeField string `json:"totallyDifferent but same underlying type"` // TODO does not yet work if equal with above
}

// Func is also annotated
// @MyFunc("see here")
func (a AnnotatedStruct) Func() string {
	return a.SomeField
}

// A Repo is a domain driven firewall into the persistence layer.
// @ee.Repo("entity")
type Repo interface {
	// GetAll returns everything
	// @ee.sql("SELECT * from xy")
	GetAll(offset int) ([]AnnotatedStruct, error)
}
