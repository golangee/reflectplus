package stuff

import (
	"github.com/golangee/uuid"
)

type MyFunc func()

type MyFunc2 func(myInt, a MyInt) (MyFunc, error)

// MyInt is cool
type MyInt int

type MyString string

// MyAlias doc
type MyAlias = MyString

type MyHopString MyString

type MyHopHopString MyHopString

// MyStruct doc
type MyStruct struct {
	// Text doc
	Text MyString // no doc
	// secret doc
	secret MyAlias `json:"myName,omitempty"`

	Id uuid.UUID

	Blub int

	What struct{ ABC string }
}

// NewMyStruct is a constructor
func NewMyStruct() *MyStruct {
	return nil
}

// SomeMethod0 doc
func (s *MyStruct) SomeMethod0() {

}

// SomeIface doc
type SomeIface interface {
	// SomeMethod0 doc
	SomeMethod0()
}

// MyInterface doc
type MyInterface interface {
	SomeIface
	SomeMethod1() MyStruct
	// SomeMethod2 doc
	SomeMethod2(a int, other string) MyStruct
	SomeMethod3(a, b, c int, other string) (MyStruct, error)
	SomeMethod4(a, b, c int, other string) ([]MyStruct, error)
	SomeMethod5(a, b, c int, other string) ([]*MyStruct, error)
	SomeMethod6(a, b, c int, other string) ([]**MyStruct, error)
	SomeMethod7(a, b, c int, other string) (*[]*MyStruct, error)
	SomeMethod8(a, b, c int, other string) ([]MyStruct, error, string)
	SomeMethod9(a, b, c int, other string) ([3]MyStruct, error)
	SomeMethod10(a, b, c int, other string) (map[string]MyStruct, error)
	SomeMethod11(a, b, c int, other string) (chan MyStruct, error)
}

type MyArray [3]byte
type MyArray2 [4]MyStruct
type MySlice []MyStruct

type MySlice2 []struct {
	a int
}

type MyChannel chan MyStruct
type MyMap map[string]*MyStruct

func SomePublicFunc() error {
	return nil
}

type RecursiveType struct {
	CanYou *RecursiveType
}

type HiddenRecursiveType struct {
	CanYou Other
}

type Other interface {
	Get() HiddenRecursiveType
}

type A []*B

type B = A

type X []X