package reflectplus

import "encoding/json"

var packages []Package

func AddPackage(pkg Package) {
	packages = append(packages, pkg)
}

func ImportMetaData(jsn []byte) (Package, error) {
	pkg := Package{}
	err := json.Unmarshal(jsn, &pkg)
	if err != nil {
		return pkg, err
	}
	AddPackage(pkg)
	return pkg, nil
}

func Packages() []Package {
	return packages
}

func FindInterface(importPath string, name string) *Interface {
	for _, p := range packages {
		for _, iface := range p.AllInterfaces() {
			if iface.ImportPath == importPath && iface.Name == name {
				return &iface
			}
		}
	}
	return nil
}

func FindPackage(importPath string) *Package {
	for _, p := range packages {
		if p.ImportPath == importPath {
			return &p
		}
		var r *Package
		p.VisitPackages(func(pkg Package) bool {
			if pkg.ImportPath == importPath {
				r = &pkg
				return false
			}
			return true
		})
		if r != nil {
			return r
		}
	}
	return nil
}
