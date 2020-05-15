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
Well it depends, the generated reflection metadata is always included, if the maintainer has included the *reflectplus*
dependency. You also have to ensure to load the package (currently only the module import path) containing the
init method which registers the additional reflection information and the interface proxy factories.