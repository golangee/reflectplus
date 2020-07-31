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

// A Qualifier consists usually of 3 parts and is separated by | (pipe). This character is disallowed on windows
// and discouraged on Unix for filenames and not allowed in URLs. It also encodes nicely in json.
// If just two | are contained, only a package is referenced. If no pipe is found, only an import path is denoted.
// Universe types have both, an empty import and an empty package name.
// Examples:
//  * github.com/me/myproject|pkgname|MyType
//  * ||int
//  * unsafe|unsafe|Pointer
//  * github.com/me/myproject|pkgname
//  * github.com/me/myproject
//
// The concept of different import paths and package names is very go-like and quite obscure, because
// you can never deduce the actual package name from its path, however it allows at least elegant solutions like this:
//  * resolving the actual dependency just by import, not from build or other environment files (well, that is not true anymore if using go.mod)
//  * versioned packages, e.g. github.com/myproject/myapi and github.com/myproject/myapi/v2 should both be named myapi
//  * using path with chars which are illegal identifiers, like github.com/myproject/my-api?may=happen#or=not (could still be myapi)
type Qualifier string

// NewQualifier assembles a new qualifier from the given parts
func NewQualifier(path, name, identifier string) Qualifier {
	return Qualifier(path + "|" + name + "|" + identifier)
}

// Path returns the import path, the first segment or the entire string, if no separators.
func (q Qualifier) Path() string {
	for i, r := range q {
		if r == '|' {
			return string(q[:i])
		}
	}

	return string(q)
}

// Name returns the actual imported package name, which is technically independent from the import path.
func (q Qualifier) Name() string {
	firstSeg := -1
	for i, r := range q {
		if r == '|' {
			if firstSeg == -1 {
				firstSeg = i + 1
			} else {
				return string(q[firstSeg:i])
			}
		}
	}

	if firstSeg != -1 {
		return string(q[firstSeg:])
	}

	return ""
}

// Identifier returns the named element from the package, which is either a constant, a variable, a function or
// a declared type.
func (q Qualifier) Identifier() string {
	pipeCount := 0
	for _, r := range q {
		if r == '|' {
			pipeCount++
		}
	}

	if pipeCount < 2 {
		return ""
	}

	for i := len(q) - 1; i >= 0; i-- {
		r := q[i]
		if r == '|' {
			return string(q[i+1:])
		}
	}
	return ""
}

// String just returns the pipe encoded qualifier
func (q Qualifier) String() string {
	return string(q)
}
