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

import (
	"fmt"
	"github.com/golangee/reflectplus/meta"
	"github.com/golangee/reflectplus/src"
	"testing"
)

func TestNewProject(t *testing.T) {
	opts := Options{}
	//mods, err := NewProject(opts, "/Users/tschinke/git/github.com/worldiety/mercurius/", nil)
	mod, err := NewProject(opts, "/Users/tschinke/git/github.com/golangee/reflectplus/internal/test", nil)
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	//fmt.Printf("%s\n", prj.table.String())

	mod.ForEachTypeAnnotation("ee.Repo", func(a meta.Annotation, named *meta.Named) {
		mod.NewType(func(t *src.TypeBuilder) {
			t.SetDoc("... is a generated implementation of a MySQL repository.\nThe name is autofilled for your\nconvenience:\n   a pre-formatted text.").
				SetName("MySQL" + named.Name + "Impl").
				AddField(func(f *src.FieldBuilder) {
					f.SetName("dbx").SetType(".int")
				}).
				AddField(func(f *src.FieldBuilder) {
					f.SetDoc("...contains things about\nstuff.")
					f.SetName("Stuff").SetType("github.com/worldiety/golangee/test.Blub")
				}).
				AddMethod(func(f *src.FuncBuilder) {
					f.SetDoc("...makes some noise.")
					f.SetName("DoJob")
				})

		})
	})

}
