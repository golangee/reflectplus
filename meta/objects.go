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

import "strconv"

type Location string

func NewLocation(filename string, line, col int) Location {
	return Location(filename + ":" + strconv.Itoa(line) + ":" + strconv.Itoa(col))
}

// A PackageQualifier consists of an import path and the according package name. This is rather obscure, because
// you can never deduce the actual package name from its path, however it allows at least elegant solutions like this:
//  * versioned packages, e.g. github.com/myproject/myapi and github.com/myproject/myapi/v2 should both be named myapi
//  * using path with chars which are illegal identifiers, like github.com/myproject/my-api (could still be myapi)
type PackageQualifier struct {
	// Path is a / separated name to resolve a package.
	Path string
	// Name is the actual identifier of the package. This is also used by default, if not renamed at import.
	Name string
}

type Annotation struct {
	Pos    Location
	Doc    string
	Name   string
	Values map[string]interface{}
}
