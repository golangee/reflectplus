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

package meta

import "testing"

func TestQualifier_Path(t *testing.T) {
	type test struct {
		In                     Qualifier
		Path, Name, Identifier string
	}

	table := []test{
		{"abc", "abc", "", ""},
		{"a|b", "a", "b", ""},
		{"a|b|c", "a", "b", "c"},
	}

	for _, r := range table {
		if r.In.Path() != r.Path {
			t.Fatalf("%s: expected %s but got %s", r.In, r.Name, r.In.Path())
		}

		if r.In.Name() != r.Name {
			t.Fatalf("%s: expected %s but got %s", r.In, r.Name, r.In.Name())
		}

		if r.In.Identifier() != r.Identifier {
			t.Fatalf("%s: expected %s but got %s", r.In, r.Identifier, r.In.Identifier())
		}
	}

}
