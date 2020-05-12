// parser contains the glue code for different parsers (e.g. go ast and go doc) and custom things like annotations.
package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

type Package struct {
	fset     *token.FileSet // only the root package has a non-nil fset
	dirname  string
	parent   *Package // the root package has a nil parent
	files    []*File
	packages []*Package
	name     string
	modName  string
	doc      *doc.Package
}

func ParsePackage(parent *Package, dirname string) (*Package, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	p := &Package{parent: parent, dirname: dirname}
	if parent == nil {
		p.fset = token.NewFileSet()
		modInfo, err := goList(dirname)
		if err != nil {
			return nil, err
		}
		if len(modInfo.ImportPath) == 0 {
			return nil, fmt.Errorf("import path for package '%s' is empty", dirname)
		}
		p.modName = modInfo.ImportPath
	}
	lastPackageName := ""
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") && file.Mode().IsRegular() {
			if file.Name() == "reflect.gen.go" {
				// do not eat our own dog food, that would cause generated headaches
				continue
			}
			fname := filepath.Join(dirname, file.Name())
			srcFile, err := ParseFile(p, fname)
			if err != nil {
				return nil, fmt.Errorf("failed to parse go file %s: %w", fname, err)
			}
			p.files = append(p.files, srcFile)
			if srcFile.node.Name.Name != lastPackageName && lastPackageName != "" {
				return nil, fmt.Errorf("package in '%s' contains ambigous package declaration '%s' and '%s'", fname, lastPackageName, srcFile.node.Name.Name)
			}
			lastPackageName = srcFile.node.Name.Name
		}

		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			fname := filepath.Join(dirname, file.Name())
			pck, err := ParsePackage(p, fname)
			if err != nil {
				return nil, fmt.Errorf("failed to parse go package %s: %w", fname, err)
			}
			p.packages = append(p.packages, pck)
		}
	}
	p.name = lastPackageName
	if p.name == "" {
		p.name = filepath.Base(p.dirname) // it may be empty, failing it not really an option, may have valid sub folders
	}

	docPkg, err := doc.NewFromFiles(p.FileSet(), p.Nodes(), p.ImportPath())
	if err != nil {
		return nil, err
	}
	p.doc = docPkg
	return p, nil
}

func (p *Package) Packages() []*Package {
	return p.packages
}

func (p *Package) Doc() *doc.Package {
	return p.doc
}

func (p *Package) TypeDoc(name string) *doc.Type {
	for _, typ := range p.doc.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

func (p *Package) FuncDoc(name string) *doc.Func {
	for _, f := range p.doc.Funcs {
		if f.Name == name {
			return f
		}
	}

	// look at constructors
	for _, typ := range p.doc.Types {
		for _, f := range typ.Funcs {
			if f.Name == name {
				return f
			}
		}
	}
	return nil
}

func (p *Package) VisitPackages(f func(pkg *Package) error) error {
	if err := f(p); err != nil {
		return err
	}
	for _, pkg := range p.packages {
		err := pkg.VisitPackages(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) VisitFiles(f func(file *File) error) error {
	for _, file := range p.files {
		err := f(file)
		if err != nil {
			return err
		}
	}
	for _, pkg := range p.packages {
		err := pkg.VisitFiles(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) VisitDoc(f func(pkg *doc.Package) error) error {
	if err := f(p.doc); err != nil {
		return err
	}
	for _, pkg := range p.packages {
		if err := pkg.VisitDoc(f); err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) Parent() *Package {
	return p.parent
}

func (p *Package) FileSet() *token.FileSet {
	if p.fset != nil {
		return p.fset
	}
	return p.Parent().FileSet()
}

func (p *Package) Files() []*File {
	return p.files
}

func (p *Package) Nodes() []*ast.File {
	var res []*ast.File
	for _, file := range p.files {
		res = append(res, file.node)
	}
	return res
}

func (p *Package) ImportPath() string {
	if p.parent == nil {
		return p.modName
	}
	return p.Parent().ImportPath() + "/" + p.Name()
}

func (p *Package) Name() string {
	return p.name
}

func (p *Package) ModName() string {
	if p.parent != nil {
		return p.Parent().ModName()
	}
	panic("unreachable")
}

// A File contains all imports and macros together
type File struct {
	fileName string    // the original file name
	node     *ast.File // the parsed AST
	parent   *Package
}

func ParseFile(parent *Package, fname string) (*File, error) {
	f := &File{parent: parent, fileName: fname}
	node, err := parser.ParseFile(parent.FileSet(), f.fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse '%s': %w", fname, err)
	}
	f.node = node
	return f, nil
}

func (f *File) Filename() string {
	return f.fileName
}

func (f *File) ResolveIdentifierImportName(name string) string {
	if len(name) == 0 {
		return ""
	}

	// there are no uppercase go types
	if unicode.IsUpper(rune(name[0])) {
		return f.Parent().ImportPath()
	}

	// TODO because we do not support private types in exported types, we just assume stdlib type
	return ""
}

func (f *File) ResolveImportName(name string) string {
	for _, spec := range f.node.Imports {
		path := trimQuotation(spec.Path.Value)
		if spec.Name != nil {
			if spec.Name.Name == name {
				return path
			}
		} else {
			if strings.HasSuffix(path, name) {
				return path
			}
		}
	}
	return name
}

func trimQuotation(str string) string {
	if strings.HasPrefix(str, `"`) {
		str = str[1:]
	}
	if strings.HasSuffix(str, `"`) {
		str = str[0 : len(str)-1]
	}
	return str
}

func (f *File) Node() *ast.File {
	return f.node
}

func (f *File) Parent() *Package {
	return f.parent
}

func (f *File) String() string {
	sb := &strings.Builder{}
	sb.WriteString(f.fileName)
	sb.WriteString(":\n")
	return sb.String()
}

type goModInfo struct {
	ImportPath string
}

func goList(dir string) (goModInfo, error) {
	cmd := exec.Command("go", "list", "-e", "-json")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	b, err := cmd.CombinedOutput()
	if err != nil {
		return goModInfo{}, fmt.Errorf("unable to get go module: %w", err)
	}
	var res goModInfo
	err = json.Unmarshal(b, &res)
	if err != nil {
		return goModInfo{}, fmt.Errorf("unable to parse 'go list' result ('%s'):%w", string(b), err)
	}
	return res, nil
}
