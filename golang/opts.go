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

package golang

// Options for the reflectplus parser
type Options struct {
	// Dir is the directory in which to run the build system's query tool that provides information about the packages.
	// If Dir is empty, the tool is run in the current directory.
	Dir string

	// Patterns contains the root packages to parse, e.g. github.com/golangee/...
	Patterns []string
}
