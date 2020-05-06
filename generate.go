package reflectplus

import (
	"fmt"
	"github.com/worldiety/reflectplus/parser"
	"os"
	"path/filepath"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Generate(dir string) error {
	if !strings.HasPrefix(dir, string(os.PathSeparator)) {
		absDir, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = filepath.Join(absDir, dir)
	}

	fmt.Printf("scanning %s\n", dir)

	pkg, err := parser.ParsePackage(nil, dir)
	if err != nil {
		return err
	}

	metaPkg, err := ParseMetaModel(pkg)
	if err != nil {
		return err
	}
	fmt.Println(metaPkg)
	return nil
}
