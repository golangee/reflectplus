package src

import "strings"

func emitDoc(w Writer, name, doc string) {
	if len(doc) > 0 {
		tmp := &strings.Builder{}
		if strings.HasPrefix(doc, "...") {
			tmp.WriteString(name)
			tmp.WriteString(" ")
			tmp.WriteString(strings.TrimSpace(doc[3:]))
		} else {
			tmp.WriteString(doc)
		}
		str := tmp.String()
		for _, line := range strings.Split(str, "\n") {
			w.Printf("// %s\n", line)
		}
	}
}
