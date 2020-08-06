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

// reflectplus generates and embedds the missing pieces in the go reflection system.
package reflectplus

import (
	"github.com/golangee/reflectplus/golang"
	"github.com/golangee/reflectplus/internal/mod"
	"os"
)

// Parse loads initiates the tooling in the given folders and loads and parses all given paths from the pattern
// including all related dependencies, including declaration from the standard library.
func Parse(opts golang.Options) (*golang.Project, error) {
	return golang.NewProject(opts)
}

// ParseModule can be invoked from any subdirectory within a valid go module and parses the module including all
// of its dependencies.
func ParseModule() (*golang.Project, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	modules, err := mod.List(dir)
	if err != nil {
		return nil, err
	}

	var rootDir string
	var patterns []string
	for _, module := range modules {
		if module.Main {
			rootDir = module.Dir
		}
		patterns = append(patterns, module.Path)
	}

	return Parse(golang.Options{
		Dir:      rootDir,
		Patterns: patterns,
	})
}
