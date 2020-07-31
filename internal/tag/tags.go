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

// Package tag provides a way to parse conventionally structured struct tags.
package tag

import (
	"sort"
	"strconv"
	"strings"
)

type Tags []StructTag

// Names returns a sorted list of unique names.
func (t Tags) Names() []string {
	set := map[string]string{}
	for _, s := range t {
		set[s.Name] = ""
	}

	res := make([]string, 0, len(set))

	for k := range set {
		res = append(res, k)
	}

	sort.Strings(res)

	return res
}

// ByName returns the index of the first StructTag or -1
func (t Tags) ByName(name string) int {
	for i, v := range t {
		if v.Name == name {
			return i
		}
	}

	return -1
}

// A StructTag represents a conventional go struct tag, as specified by reflect.StructTag. Keep in mind
// that this is just a convention which is not enforced by the compiler and the spec only requires an
// arbitrary string literal (see https://golang.org/ref/spec#Struct_types).
type StructTag struct {
	// Name of the tag. This must not be unique.
	Name string
	// Values are comma separated
	Values []string
}

// Parse inspects the given string and returns a slice of StructTag. Because the specification does not enforce
// a specific format, we will never return an error. Unparseable struct tags are silently ignored. The comma
// separated values are trimmed.
func Parse(tagStr string) Tags {
	if strings.HasPrefix(tagStr, "`") {
		tagStr = tagStr[1:]
	}

	if strings.HasSuffix(tagStr, "`") {
		tagStr = tagStr[:len(tagStr)-1]
	}
	tagStr = strings.TrimSpace(tagStr)
	return mapFieldTags(tagStr)
}

// based on reflect/type.go StructTag.Lookup
func mapFieldTags(tag string) []StructTag {
	var res []StructTag

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil // does this ever happen with the check above?
		}

		//dunno why this has a broken name, the original code seems to have this as a defacto-bug
		if strings.HasPrefix(name, ",") {
			name = name[1:]
		}

		tag := StructTag{
			Name:   name,
			Values: strings.Split(value, ","),
		}

		for i, v := range tag.Values {
			tag.Values[i] = strings.TrimSpace(v)
		}

		res = append(res, tag)
	}

	return res
}
