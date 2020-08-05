package golang

import (
	"github.com/golangee/reflectplus/meta"
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
