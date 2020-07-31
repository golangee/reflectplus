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

package tag

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		args string
		want Tags
	}{
		{"", nil},
		{"a:b", nil}, // no quotes: syntax error
		{`a:"b"`, []StructTag{{"a", []string{"b"}}}},
		{"`a:\"b\"`", []StructTag{{"a", []string{"b"}}}},
		{`a:"b,c"`, []StructTag{{"a", []string{"b", "c"}}}},
		{`a:"b, c"`, []StructTag{{"a", []string{"b", "c"}}}},
		{` a:" b , c "`, []StructTag{{"a", []string{"b", "c"}}}},
		{` a: " b , c "`, nil},    // a space after a colon: syntax error
		{` a:" b , c `, nil},      // unbalanced quote: syntax error
		{"a:'\u2639\u2639'", nil}, // invalid quote: syntax error
		{` a:" , c "`, []StructTag{{"a", []string{"", "c"}}}},
		{`json:"a" xml:"b"`,
			[]StructTag{
				{"json", []string{"a"}},
				{"xml", []string{"b"}},
			},
		},
		{`	json:"a " xml:"	b "`,
			[]StructTag{
				{"json", []string{"a"}},
				{"xml", []string{"b"}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := Parse(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
