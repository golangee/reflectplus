# reflectplus [![GoDoc](https://godoc.org/github.com/golangee/reflectplus?status.svg)](http://godoc.org/github.com/golangee/reflectplus)
The missing reflection bits for go. This library parses your go source code and generates
reflection information at compile time, which can be inspected later at runtime. This can be also used
for code generation. 

Using this library, you can work around the following issues:
* inspect function parameter names: https://github.com/golang/go/issues/12384
* create interface proxy at runtime: https://github.com/golang/go/issues/16522 and https://github.com/golang/go/issues/4146
* annotation support (comments): https://github.com/golang/go/issues/36669 and https://stackoverflow.com/questions/37488509/how-to-get-annotation-of-go-language-function
* discover package types and funcs: https://stackoverflow.com/questions/32132064/how-to-discover-all-package-types-at-runtime
* get reflect.Type by name: https://stackoverflow.com/questions/40879748/golang-reflect-get-type-representation-from-name
* get method or function parameter names: https://stackoverflow.com/questions/31377433/getting-method-parameter-names

related work:
* https://github.com/MarcGrol/golangAnnotations, but provides only a hard coded set of annotations and
is not module ready.
* https://github.com/cosmos72/gomacro, fancy but does not provide go type information.
* go-doc parser: https://golang.org/pkg/go/doc/ (is used)
* go-ast parser: https://golang.org/pkg/go/ast/ (is used)

## roadmap
- [x] interfaces
- [x] structs
- [x] package level functions
- [x] annotations
- [x] keep comments
- [x] struct constructors
- [ ] annotation validation at parsing time
- [ ] package level variables
- [ ] package level constants
- [ ] type aliases
- [ ] other type definitions
- [x] interface proxy (stub code generation)
- [ ] private functions, methods, types (will never be supported)
- [x] multiline annotation values

## annotation support
In contrast to macros, annotations are just passive data key/value pairs in JSON notation for any 
type or function. The following notations are allowed:
 
```go
// A MyRepo is for ...
// @Repo
// @Repo()
// @Repo({}) // comments allowed, outer {} can be omitted 
// @Repo({"value":5})
// @Repo(5) // implicitly wrapped into {"value": 5}
// @Repo("text") // implicitly wrapped into {"value": "text"}
// @Repo("value":"te:xt") // this is fine 
// @Repo("values":["can","be","multiple"])
// @Repo("anyKey":"anyValue","num":5,"bool":true,"nested":{"care":{"of":["your", "head"]}})
// @Repo("""
//    {
//      "json":"front matter"
//    }
//    this is 
//    a multiline string 
//    or json literal.
//    However line breaks and additional start/ending spaces are discarded and replaced by 
//    a single space.
// """)
type MyRepo interface{
    //...
}
```

## interface proxy support
The reflection and proxy support is just as you would have expected it:

```go
package main

import (
    "my/module/pckage"
    "fmt"
    "github.com/golangee/reflectplus"
    _ "my/module"
)

func main(){
    iface := reflectplus.FindInterface("my/module/pckage","MyInterface")
    fmt.Println(iface)

    proxy, err := reflectplus.NewProxy("my/module/pckage", "MyInterface", func(method string, args ...interface{}) []interface{} {
        fmt.Printf("hello %s\n", method)
        return nil
    })
    if err != nil {
        panic(err)
    }
    proxy.(pckage.MyInterface).MyMethod()
}


```

## usage

### go generate (recommended)
```bash
# create a file like my/module/cmd/gen/gen.go
//go:generate go run gen.go
package main

import (
	"github.com/golangee/reflectplus"
)

func main() {
	reflectplus.Must(reflectplus.Generate("../.."))
}

# import dependency
go get github.com/golangee/reflectplus

# go generate
go generate ./...

# the generated file is my/module/reflect.gen.go, you need to import it, to run its init method
# e.g. in my/module/cmd/app/main.go
package main

import _ "github.com/my/package"

func main(){
//...
}
```

### standalone
```bash
GO111MODULE=off && go get -u github.com/golangee/reflectplus/cmd/reflectplus
cd my/module
reflectplus
```

## FAQ
### Does it work in go path?
This is a legacy configuration and not tested, so probably not.

### Does it work with multiple modules?
TODO: this will be currently changed, and you only use it in your top most project through
go generate

Well it depends, the generated reflection metadata is always included, if the maintainer has included the *reflectplus*
dependency. You also have to ensure to load the package (currently only the module import path) containing the
init method which registers the additional reflection information and the interface proxy factories.


## nomenclature and the go type system
The [specification](https://golang.org/ref/spec#Types) defines a type as follows
>A type determines a set of values together with operations and methods specific to those 
>values. A type may be denoted by a type name, if it has one, or specified using a type 
>literal, which composes a type from existing types.
>
>The language predeclares certain type names. Others are introduced with type declarations. 
>Composite types—array, struct, pointer, function, interface, slice, map, and channel types—may 
>be constructed using type literals.
>
>Each type T has an underlying type: If T is one of the predeclared boolean, numeric, or string 
>types, or a type literal, the corresponding underlying type is T itself. Otherwise, 
>T's underlying type is the underlying type of the type to which T refers in its type declaration.

So, this is the prerequisite to an actual [*type declaration*](https://golang.org/ref/spec#Type_declarations)
>A type declaration binds an identifier, the type name, to a type. Type declarations come in two forms: 
>alias declarations and type definitions.

A [type definition](https://golang.org/ref/spec#Type_declarations) is defined as follows
>A type definition creates a new, distinct type with the same underlying type and operations 
>as the given type, and binds an identifier to it.
>The new type is called a defined type. It is different from any other type, including the type 
>it is created from.

### examples
The following sub chapters show some examples definitions and how they are represented.

#### basic type 1
```go
type MyInt int
```
* represented as `*ast.TypeSpec` or `go/types.Named`
* kind: type declaration
* type name: MyInt
* Underlying type: `*go/types.Basic(int)`

#### basic type 2
```go
type MyStr string
type MyOtherString MyInt
```
* represented as `*ast.TypeSpec` or `go/types.Named`
* kind: type declaration
* type name: MyOtherString
* Underlying type: `*go/types.Basic(string)`

#### struct
```go
type MyStruct struct {
	Text MyString 
	secret MyAlias
	Id uuid.UUID
}

func (s *MyStruct) SomeMethod0() {}
```
* represented as `*ast.TypeSpec` or `go/types.Named`
* kind: type declaration
* type name: MyStruct
* Underlying type: `*go/types.Struct`
  * fields: `[]*go/types.Var` (providing name and recursive type reference)
  * tags: `[]string
  * methods: `[]*go/types.Func` (providing name)
    * Type: *go/types.Signature
      
### what we've learned
A type declaration has a name, and a reference to its (unnamed) underlying type. This tupel declares
always a unique type definition. The underlying
type is used by the [*type conversion* system](https://golang.org/ref/spec#Conversions). 
If a conversion only changes the type and not its representation, no runtime cost is involved. 
Note that method declarations on a type do not belong to the underlying type, but just to the
*defined type* (remember a *declared type* is either an *alias declaration* or a *type definition*,
but an alias cannot carry methods).

There is no inheritance or whatsoever involved. Types always only carry their most basic underlying type, 
independent of how many indirections are made in the declaration. This only ensures the possibility to
allow type conversions. This shows also why the conversion of e.g. slices or array of different
types cannot be done, because they each form a distinct underlying type.

It is still unclear if and how a future generic specification fits into. As it currently
stands out, custom generics (just like built-in generics today) have a fixed ordered semantic
for the according type parameter, which itself are either *defined types* or even
anonymous type declarations. But probably they create a new underlying type, just as today
with the build-ins. 

It looks like anonymous types are actually equivalent to underlying
types.

There is no inheritance in Go and the compiler and resolver do not even keep the information
about chained type hierarchies. This is only kept internally to check for recursive type
definitions. The only available information is the final underlying type which is never a *named
typed*, hence not containing any positional information: the underlying type is always an abstract
concept and forms the central part of the *ducktyping* logic in Go.

### design decisions


#### representing syntactical inheritance
We do not introduce an artificial inheritance regarding the syntactical declared type hierarchy 
in Go because it has no defined semantic meaning. Even if this information resides in the AST
we cannot access it using *golang/x/tools* because the resolved and parsed type information is 
at best available in a private field (*types.Named.orig*) which is only used for recursion detection
and its content is no further specified and probably subject to change. We do not want to use 
*unsafe* trickery in our model to promise something we cannot keep. 

The benefit of inherited type annotations is probably not worth the hassle and headaches we may
otherwise introduce. A better substitute would be to create a custom annotation which itself
allows importing annotations from other locations.

```
┌───────────────────────────┐            ┌──────────────────┐         
│   type OtherThing MyInt   │            │  type MyInt int  │         
└─────────────┬─────────────┘            └─────────┬────────┘         
              │                                    │                  
              │ underlying type                    │ underlying type  
              │                                    │                  
           ┌──▼──┐                              ┌──▼──┐               
           │ int │                              │ int │               
           └─────┘                              └─────┘               
```

#### duplication of interface methods
We keep redundant method signatures in the underlying type and concrete method definitions in
the named type. Because each interface, and their corresponding methods may have their own 
unique documentation, which we want to process, it is clearer to introduce a clean separation. 

```
       ┌───────────────────────────────────────┐          
       │// MyInterface Doc                     │          
       │type MyInterface interface {           │          
       │   // MyMethod Doc                     │          
       │   MyMethod()                          │          
       │}                                      │          
       └──┬────────────────────────────┬───────┘          
          │                            │                  
          │                            │ underlying type  
          │                            │                  
┌─────────▼──────────┐               ┌─▼───────────┐      
│ Declared Interface │               │  Interface  │      
└─────────────┬──────┘               └──────┬──────┘      
          ┌───▼────────────────┐          ┌─▼───────────┐ 
          │  Declared Methods  │          │ Signatures  │ 
          └────────────────────┘          └─────────────┘ 
```

An underlying type also never carries file and positional information, which are unique per
named type instead. Also struct tags are omitted from the underlying type, and instead annotated
in the named type, as defined by language specification. See also 
[Type Identity](https://golang.org/ref/spec#Type_identity).


#### mixture of underlying and named types
We do not mix them, because it causes a lot of headache and is wrong anyway.
Even if anonymous types look exactly the same as their underlying type, they
are different in a way that anonymous types have at least their 
own source location. At the end it is probably more a kind of 
compiler sugar, to avoid explicit type casts or just named types without
a name.

```go
type MyType struct{
    MyField struct{ 
      OtherField int
    }   
}

func MyFunc(params struct{MyField int}, iface interface{Do()}){}
```