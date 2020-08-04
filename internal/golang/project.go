package golang

import (
	"fmt"
	"github.com/golangee/reflectplus/meta"
	"github.com/golangee/reflectplus/src"
)

type Project struct {
	table *meta.Table
}

func (p *Project) ForEachTypeAnnotation(annotationName string, f func(a meta.Annotation, named *meta.Named)) {
	for _, v := range p.table.Declarations {
		if v.Named != nil {
			for _, a := range v.Named.Annotations {
				if a.Name == annotationName {
					f(a, v.Named)
				}
			}
		}
	}
}

func (p *Project) NewType(f func(t *src.TypeBuilder)) {
	file := src.NewFileBuilder()
	file.NewType(f)

	w := &src.BufferedWriter{}
	file.Emit(w)
	str, err := w.Format()
	fmt.Println(str)
	if err != nil {
		fmt.Println(err)
	}
}
