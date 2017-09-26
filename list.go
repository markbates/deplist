package deplist

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var commonSkips = []string{"vendor", ".git", "examples", "node_modules"}

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
		wg.Go(func() error {
			return l.process(path, info)
		})
		return nil
	})

	if err != nil {
		return l.deps, errors.WithStack(err)
	}

	err = wg.Wait()

	return l.deps, err
}

func (l *lister) add(dep string) {
	l.moot.Lock()
	defer l.moot.Unlock()
	l.deps[dep] = dep
}

func (l *lister) process(path string, info os.FileInfo) error {
	path = strings.TrimPrefix(path, l.root)
	if info.IsDir() {

		for _, s := range l.skips {
			if strings.Contains(strings.ToLower(path), s) {
				return nil
			}
		}

		cmd := exec.Command("go", "list", "-e", "-f", `'* {{ join .Deps  "\n"}}'`, "./"+path)
		b, err := cmd.Output()
		if err != nil {
			fmt.Println(string(b))
			return errors.WithStack(err)
		}

		list := strings.Split(string(b), "\n")

		for _, g := range list {
			if strings.Contains(g, "github.com") || strings.Contains(g, "bitbucket.org") {
				g = strings.TrimPrefix(g, "'* ")
				l.add(g)
			}
		}
	}
	return nil
}
