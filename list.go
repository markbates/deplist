package deplist

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/markbates/oncer"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var commonSkips = []string{"vendor", ".git", "examples", "node_modules", ".idea"}

type lister struct {
	root  string
	skips []string
	deps  map[string]string
	moot  *sync.Mutex
}

func List(skips ...string) (map[string]string, error) {
	pwd, _ := os.Getwd()
	l := lister{
		root:  pwd,
		skips: append(skips, commonSkips...),
		deps:  map[string]string{},
		moot:  &sync.Mutex{},
	}
	wg := &errgroup.Group{}

	err := filepath.Walk(pwd, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return l.process(path, info)
		}
		return nil
	})

	if err != nil {
		return l.deps, errors.WithStack(err)
	}

	err = wg.Wait()

	return l.deps, err
}

func (l *lister) add(dep string) {
	if dep == "." || dep == "C" {
		return
	}
	l.moot.Lock()
	defer l.moot.Unlock()
	l.deps[dep] = dep
}

func (l *lister) process(path string, info os.FileInfo) error {
	path = strings.TrimPrefix(path, l.root)
	path = strings.TrimPrefix(path, string(filepath.Separator))
	if info.IsDir() {

		for _, s := range l.skips {
			if strings.Contains(strings.ToLower(path), s) {
				return nil
			}
		}

		return l.find(".", filepath.Dir(path))
	}
	return nil
}

func (l *lister) find(name string, dir string) error {
	var err error
	oncer.Do(name+dir, func() {
		ctx := build.Default

		pkg, err := ctx.Import(name, dir, 0)

		if err != nil {
			if !strings.Contains(err.Error(), "cannot find package") {
				if _, ok := errors.Cause(err).(*build.NoGoError); !ok {
					err = errors.WithStack(err)
					return
				}
			}
		}

		if pkg.Goroot {
			err = nil
			return
		}

		imps := append(pkg.Imports, pkg.TestImports...)
		for _, imp := range imps {
			if err := l.find(imp, pkg.Dir); err != nil {
				err = errors.WithStack(err)
				return
			}
		}
		l.add(pkg.ImportPath)
	})
	return err
}
