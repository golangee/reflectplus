package src

import (
	"fmt"
	"go/format"
	"strings"
)

type Writer interface {
	Printf(str string, args ...interface{})
}

type BufferedWriter strings.Builder

func (b *BufferedWriter) Printf(format string, args ...interface{}) {
	if len(args) == 0 {
		(*strings.Builder)(b).WriteString(format)
		return
	}

	(*strings.Builder)(b).WriteString(fmt.Sprintf(format, args...))
}

func (b *BufferedWriter) String() string {
	return (*strings.Builder)(b).String()
}

func (b *BufferedWriter) Format() (string, error) {
	buf, err := format.Source([]byte(b.String()))
	if err != nil {
		return b.WithLinesNumbers(), err
	}

	return string(buf), nil
}

func (b *BufferedWriter) WithLinesNumbers() string {
	sb := &strings.Builder{}
	for i, line := range strings.Split(b.String(), "\n") {
		sb.WriteString(fmt.Sprintf("%4d: %s\n", i+1, line))
	}
	return sb.String()
}
