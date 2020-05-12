package parser

import (
	"encoding/json"
	"strings"
)

// An AnnotationParserError means that an annotation has been found but could not be parsed
type AnnotationParserError struct {
	Text    string
	Details string
}

func (a *AnnotationParserError) Error() string {
	return "ParserError: " + a.Text + ": " + a.Details
}

//A NoAnnotationError means that Text did not contain any annotation
type NoAnnotationError struct {
	Text string
}

func (a NoAnnotationError) Error() string {
	return "Not an annotation: " + a.Text
}

func IsNoAnnotationError(err error) bool {
	_, ok := err.(NoAnnotationError)
	return ok
}

// An Annotation is actually an @-prefixed-named json object one-liner
type Annotation struct {
	Doc    string
	Text   string
	Name   string
	Values map[string]interface{}
}

func validDotIdentifier(str string) bool {
	if len(str) == 0 {
		return false
	}

	if str[0] == '.' || (str[0] >= '0' && str[0] <= '9') {
		return false
	}

	for _, c := range str {
		validChar := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '.'
		if !validChar {
			return false
		}
	}
	return true
}

// ParseAnnotation is a simple minimal parser which evaluates a single line as an Annotation.
// If the given line is not an annotation at all, *NoAnnotationError is returned.
// An annotation always starts with an @ followed by a char sequence with optional opening and closing braces.
//
// Examples
//  *not valid: @
//  *not valid: @()
//  *valid: @a
//  *valid: @a.b.c() // some comment or text )
//  *invalid: @a.b.c("hello") ) // this is wrong syntax
func ParseAnnotation(line string) (Annotation, error) {
	annotation := Annotation{
		Text: line,
	}
	cleanLine := line
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		cleanLine = cleanLine[0:commentIdx]
		annotation.Doc = strings.TrimSpace(line[commentIdx+2:])
	}

	cleanLine = strings.TrimSpace(cleanLine)
	if len(cleanLine) == 0 {
		return annotation, NoAnnotationError{line}
	}

	if cleanLine[0] != '@' {
		return annotation, NoAnnotationError{line}
	}

	openArg := strings.Index(cleanLine, "(")
	closeArg := strings.LastIndex(cleanLine, ")")

	if openArg != closeArg && (openArg == -1 || closeArg == -1) {
		return annotation, &AnnotationParserError{line, "unbalanced open/close argument braces"}
	}
	annotationName := ""
	// case where no params are given
	if openArg == -1 {
		annotationName = cleanLine[1:]
	} else {
		annotationName = cleanLine[1:openArg]
	}

	if !validDotIdentifier(annotationName) {
		return annotation, &AnnotationParserError{line, "annotation identifier is invalid"}
	}

	annotation.Name = annotationName

	if openArg == -1 {
		// no args
		annotation.Values = make(map[string]interface{})
		return annotation, nil
	}

	args := strings.TrimSpace(cleanLine[openArg+1 : closeArg])

	// missing {} can be detected easily
	if !strings.HasPrefix(args, "{") {
		// try 0: correct key/value but just without braces
		tmp := "{" + args + "}"
		values, err1 := parseValues(tmp, line)
		if err1 != nil {
			// try 1: just a string, without key/value
			tmp := `{"value":` + args + "}"
			values, err2 := parseValues(tmp, line)
			if err2 != nil {
				// nothing we can do
				return annotation, err1
			}
			annotation.Values = values
		} else {
			annotation.Values = values
		}
	} else {
		// otherwise just try to parse without processing
		values, err := parseValues(args, line)
		if err != nil {
			return annotation, err
		}
		annotation.Values = values
	}

	return annotation, nil
}

func parseValues(args, line string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := json.Unmarshal([]byte(args), &values)
	if err != nil {
		return nil, &AnnotationParserError{line, "annotation arguments are invalid: " + err.Error()}
	}
	return values, nil
}

// ParseAnnotations tries to parse all annotations from the given text and only returns ParserErrors
func ParseAnnotations(text string) ([]Annotation, error) {
	var res []Annotation
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		a, err := ParseAnnotation(line)
		if err != nil && !IsNoAnnotationError(err) {
			return res, err
		}
		if err == nil {
			res = append(res, a)
		}
	}
	return res, nil
}
