package src

import "strings"

// A Qualifier is a <path>.<name> string, e.g. github.com/myproject/mymod/mypath.MyType. Universe types have a
// leading dot, like .int or .float32 or .error. It does not carry any information about the actual package name,
// so it can only be used in an explicitly named import context.
type Qualifier string

func (q Qualifier) Name() string {
	i := strings.LastIndex(string(q), ".")
	return string(q[i+1:])
}

func (q Qualifier) Path() string {
	i := strings.LastIndex(string(q), ".")
	return string(q[:i])
}
