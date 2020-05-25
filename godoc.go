package reflectplus

import "strings"

// DocShortText returns the first sentence from text.
func DocShortText(text string) string {
	text = DocText(text)
	idx := strings.Index(text, ".")
	if idx < 0 {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text[0:idx])
}

// DocText returns all text until the first annotation is found (a new line starting with @)
func DocText(text string) string {
	sb := &strings.Builder{}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "@") {
			break
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}
