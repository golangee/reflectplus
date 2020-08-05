package src

import "strings"


// A Qualifier is a <path>.<name> string, e.g. github.com/myproject/mymod/mypath.MyType. Universe types have a
// leading dot (or no dot at all), like .int or .float32 or .error.
// It does not carry any information about the actual package name, so it can only be used in an explicitly named import
// context, which is sufficient per definition. Generic types cannot be expressed and must use the TypeDecl struct.
type Qualifier string

func (q Qualifier) Name() string {
	i := strings.LastIndex(string(q), ".")
	if i == -1 {
		return string(q)
	}

	return string(q[i+1:])
}

func (q Qualifier) Path() string {
	i := strings.LastIndex(string(q), ".")
	if i == -1 {
		return ""
	}

	return string(q[:i])
}
