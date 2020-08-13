package src

import (
	"strings"
	"unicode"
)

// snakeCaseToCamelCase converts strings like "my_snake-case" into "MySnakeCase".
func snakeCaseToCamelCase(str string) string {
	sb := &strings.Builder{}
	nextUp := true
	for _, r := range str {
		if r == '-' || r == '_' {
			nextUp = true
			continue
		}

		if nextUp {
			sb.WriteRune(unicode.ToUpper(r))
			nextUp = false
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
