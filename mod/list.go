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

// Package mod provides a helper to invoke go tooling and get information about the module.
package mod

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Modules is a slice of modules, with some helper functions.
type Modules []*Module

// Main returns the main modules, there is always exact one, otherwise the Modules is in an invalid state.
func (m Modules) Main() *Module {
	for _, mod := range m {
		if mod.Main {
			return mod
		}
	}

	panic("invalid modules")
}

// A Module describes an compilable and versioned part of a go project.
type Module struct {
	// Path to import this module
	Path string

	// Main is true, if this is the main module
	Main bool

	// Dir denotes the local folder where the actual source code is available
	Dir string

	// GoMod is the path to the go.mod file
	GoMod string

	// Indirect is true, if this is a transitive dependency
	Indirect bool

	// Version is a semantic version string
	Version string

	// Replace is either nil or not nil and contains overridden module information, which is either a fork
	// or a local version.
	Replace *Module

	// Error describes a module failure
	Error *ModuleError `json:",omitempty"`
}

// A ModuleError describes an error loading information about a module.
type ModuleError struct {
	Err string // error text
}

// List invokes "go list -json -m all" in the given directory and returns a flat list of all used modules.
func List(dir string) (Modules, error) {
	cmd := exec.Command("go", "list", "-json", "-m", "all")
	cmd.Dir = dir
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unable to 'go mod list -json -m all': %w", err)
	}

	// so this is not even json-lines, just a bunch of json objects, one after another, each having arbitrary amount
	// of lines, so we assemble a real json array from it
	betweenObjs := regexp.MustCompile(`}\n\s*{`)

	arr := &strings.Builder{}
	arr.WriteString("[")
	arr.WriteString(betweenObjs.ReplaceAllString(string(res), "},{"))
	arr.WriteString("]")

	r := Modules{}
	err = json.Unmarshal([]byte(arr.String()), &r)
	if err != nil {
		return nil, fmt.Errorf("unable to decode json from '%s':%w", string(res), err)
	}

	return r, nil
}
