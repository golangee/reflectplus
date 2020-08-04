package src

import "testing"

func TestQualifier_Path(t *testing.T) {
	if Qualifier(".int").Path() != "" {
		t.Fatal(Qualifier(".int").Path())
	}

	if Qualifier("github.com/myprj/mypath.MyType").Path() != "github.com/myprj/mypath" {
		t.Fatal()
	}

	if Qualifier("github.com/myprj/mypath.MyType").Name() != "MyType" {
		t.Fatal(Qualifier("github.com/myprj/mypath.MyType").Name())
	}
}
