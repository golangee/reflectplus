// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// An AnnotationParserError means that an annotation has been found but could not be parsed
type AnnotationParserError struct {
	Text    string
	LineNo  int
	Details string
}

func (a *AnnotationParserError) Error() string {
	return "ParserError: " + a.Text + ":" + strconv.Itoa(a.LineNo) + ": " + a.Details
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

// ParseAnnotations tries to parse any annotations from the given text.
func ParseAnnotations(text string) ([]Annotation, error) {
	var res []Annotation
	lines := strings.Split(text, "\n")
	for lineNo := 0; lineNo < len(lines); lineNo++ {
		line := lines[lineNo]
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "@") {
			commentIdx := strings.Index(trimmedLine, "//")
			doc := ""
			if commentIdx >= 0 {
				trimmedLine = trimmedLine[0:commentIdx]
				doc = strings.TrimSpace(line[commentIdx+2:])
			}
			openArg := strings.Index(trimmedLine, "(")
			closeArg := strings.LastIndex(trimmedLine, ")")

			multilineMarker := strings.Index(trimmedLine, `"""`)

			if openArg != closeArg && (openArg == -1 || closeArg == -1) && multilineMarker == -1 {
				return nil, &AnnotationParserError{line, lineNo, "unbalanced open/close argument braces"}
			}

			annotationName := ""
			// case where no params are given
			if openArg == -1 {
				annotationName = trimmedLine[1:]
			} else {
				annotationName = trimmedLine[1:openArg]
			}

			if !validDotIdentifier(annotationName) {
				return nil, &AnnotationParserError{line, lineNo, "annotation identifier is invalid"}
			}

			if multilineMarker == -1 {
				var args string
				if openArg > -1 {
					args = strings.TrimSpace(trimmedLine[openArg+1 : closeArg])
				}
				annotation := parseSingleLineAnnotation(line, lineNo, annotationName, args, doc)
				res = append(res, annotation)
			} else {
				// ok that's ugly, we have a multiline marker, so we will now search the eof which is another triple followed by )
				buf := &strings.Builder{}
				buf.WriteString(line[strings.Index(line, `"""`)+3:]) // use original index, without trimming and comment removal
				for {
					lineNo++
					nextLine := lines[lineNo]
					eofMarker := strings.LastIndex(nextLine, `""")`)
					if eofMarker >= 0 {
						buf.WriteString(nextLine[:eofMarker])
						break
					}
					buf.WriteString(nextLine)
					buf.WriteRune('\n')
				}
				annotation := parseMultiLineAnnotation(annotationName, buf.String())
				res = append(res, annotation)
			}

		}
	}
	return res, nil
}

// parseSingleLineAnnotation uses args and duck-types it into various format styles, it cannot fail,
// because we support many types of lax annotations and if we fail entirely, we just return the original string
// (without quotes)
//  @anno()
//  @anno("asdf") // "value":"asdf"
//  @anno(5) // "value":5
//  @anno("key":"value","o\"ther":"key") //json
//  @anno({"key":"value","o\"ther":"key"}) //json
//  @anno(any "ugly and totally un) parseable string) // "value":"any...
func parseSingleLineAnnotation(line string, lineNo int, name string, args string, doc string) Annotation {
	a := Annotation{
		Doc:    doc,
		Text:   line,
		Name:   name,
		Values: map[string]interface{}{},
	}
	args = strings.TrimSpace(args)

	// 1. be just empty
	if args == "" {
		return a
	}

	// 2. be just json
	values, err := parseJson(args, line, lineNo)
	if err != nil {
		// 3. if not, just try with omitted braces
		values, err = parseJson(fmt.Sprintf(`{%s}`, args), line, lineNo)
		if err != nil {
			// 4. if not, put it as a value "as is" and hope it is correctly json escaped
			values, err = parseJson(fmt.Sprintf(`{"value":%s}`, args), line, lineNo)
			if err != nil {
				// 5. we cannot parse it at all, so just keep it as a simple string value (but remove quotes, if any)
				if strings.HasPrefix(args, `"`) && strings.HasSuffix(args, `"`) {
					args = args[1 : len(args)-1]
				}
				values = map[string]interface{}{"value": args}
			}
		}
	}

	a.Values = values
	return a
}

// parseMultiLineAnnotation is quite similar but supports an optional json front matter. It will also never fail.
//  @anno("""
//     {
//          "front":"matter",
//          "in":"json",
//          "value":"is overridden"
//     }
// 		Any rubbish afterwards is put into the value. Keep in mind that the opening and closing braces must be
// 		each in it's own line.
//  """)
func parseMultiLineAnnotation(name string, args string) Annotation {
	a := Annotation{
		Text: args,
		Name: name,
	}

	foundOpenBrace := false
	lines := strings.Split(args, "\n")
	frontMatter := &strings.Builder{}
	body := &strings.Builder{}
	bodyStartAt := -1
	for idx, line := range lines {
		if bodyStartAt == -1 {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "{" {
				foundOpenBrace = true
			}
			if foundOpenBrace {
				frontMatter.WriteString(line)
			}
			if trimmedLine == "}" {
				bodyStartAt = idx + 1
			}
		} else {
			body.WriteString(line)
		}
	}
	// 1. if we found no frontmatter, just return raw string
	if bodyStartAt == -1 {
		a.Values = map[string]interface{}{"value": args}
		return a
	}

	// 2. if we found frontmatter, try to parse it. If we fail, just pass raw string
	values, err := parseJson(frontMatter.String(), args, 0)
	if err != nil {
		a.Values = map[string]interface{}{"value": args}
		return a
	}

	// override with body
	origVal := values["value"]
	if origVal != nil {
		values["_value"] = origVal
	}
	values["value"] = body.String()
	a.Values = values
	return a
}

func parseJson(args, line string, lineNo int) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := json.Unmarshal([]byte(args), &values)
	if err != nil {
		return nil, &AnnotationParserError{line, lineNo, "annotation arguments are invalid: " + err.Error()}
	}
	return values, nil
}

// CanonizeString removes any new lines, replaces it by a single whitespace and appends (" and ") to it
func CanonizeString(s string) string {
	s = strings.TrimSpace(s)
	sb := &strings.Builder{}
	lines := strings.Split(s, "\n")
	everWritten := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			if everWritten {
				sb.WriteRune(' ')
			}
			sb.WriteString(line)
			everWritten = true
		}
	}
	return sb.String()
}
