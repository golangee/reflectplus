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

	//fmt.Printf("%s\n", mod.table.String())

	/*
		mod.ForEachTypeAnnotation("ee.Repo", func(a meta.Annotation, named *meta.Named) {
			str := src.NewFile("gentest").
				SetGeneratorName("unittest").
				SetPackageDoc("Package gentest is here to test.").
				AddTypes(src.NewStruct("MySQL"+named.Name+"Impl").
					SetDoc("... is a generated implementation of a MySQL repository.\nThe name is autofilled for your\nconvenience:\n   a pre-formatted text.").
					AddFields(
						src.NewField("dbx", src.NewTypeDecl("int")),
						src.NewField("Stuff", src.NewTypeDecl("github.com/worldiety/golangee/test.Blub")).
							SetDoc("...contains things about\nstuff."),
						src.NewField("cache", src.NewSliceDecl(src.NewTypeDecl("string"))),
						src.NewField("lookup", src.NewMapDecl(src.NewTypeDecl("float64"), src.NewPointerDecl(src.NewTypeDecl("github.com/worldiety/golangee/test.Bla")))),
					).
					AddMethods(
						src.NewFunc("DoJob").
							SetDoc("...makes some noise.").
							SetPointerReceiver(true).
							AddParams(
								src.NewParameter("in", src.NewTypeDecl("io.Reader")),
							).
							AddResults(
								src.NewParameter("", src.NewTypeDecl("io.Writer")),
								src.NewParameter("", src.NewErrorDecl()),
							).
							AddBody(src.NewBlock().
								AddLine("var x ", src.NewTypeDecl("io.Duffer")).
								Var("y", src.NewTypeDecl("io.Duffer")).
								ForEach("", "fields", src.NewBlock().
									AddLine(src.NewTypeDecl("fmt.Println"), "(\"field\")")).
								If("true", src.NewBlock().AddLine(src.NewCallDecl("fmt.Println"), "(5)")).
								AddLine("return x,nil"),
							),
					),
				).String()
			fmt.Println(str)
		})*/

	mod.ForEachInterface(func(pkg *meta.Package, id meta.DeclId, named *meta.Named, iface *meta.Interface) {
		fmt.Println("iface ", pkg.Path, "=>", named.Name)
		impl, err := mod.Implement(id, func(ctx MethodContext) {
			fmt.Println("   ?>", ctx.MethodAnnotations)
			if len(ctx.Method.Results()) > 0 {
				ctx.Method.AddBody(src.NewBlock().
					Var("x", ctx.Method.Results()[0].Decl()))
			}

		})

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(src.NewFile("test").AddTypes(impl).String())
	})

}
