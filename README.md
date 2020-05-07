# reflectplus
The missing reflection bits for go. This library parses your go source code and generates
reflection information at compile time, which can be inspected later at runtime. This can be also used
for code generation. 

Using this library, you can work around the following issues:
* inspect function parameter names: https://github.com/golang/go/issues/12384
* create interface proxy at runtime: https://github.com/golang/go/issues/16522
* annotation support (comments): https://github.com/golang/go/issues/36669 and https://stackoverflow.com/questions/37488509/how-to-get-annotation-of-go-language-function

related work:
* https://github.com/MarcGrol/golangAnnotations, but provides only a hard coded set of annotations and
is not module ready.
* https://github.com/cosmos72/gomacro, fancy but does not provide go type information.



## how annotations work
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
// // @Repo("te:xt") invalid notation for auto detection
// @Repo("value":"te:xt") // this is fine 
// @Repo("values":["can","be","multiple"])
// @Repo("anyKey":"anyValue","num":5,"bool":true,"nested":{"care":{"of":["your", "head"]}})
type MyRepo interface{
    //...
}
```

## usage

### go generate (recommended)

### standalone