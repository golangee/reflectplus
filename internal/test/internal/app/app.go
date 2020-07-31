package app

import (
	"github.com/golangee/reflectplus/internal/test/internal/domain"
)

type MyApp struct {
	ctr *domain.ADomainController
}

func (a MyApp) Run() error {
	return a.ctr.DoJob()
}

// CrazyFunc is like this
func CrazyFunc() (int, error) {
	return 0, nil
}
