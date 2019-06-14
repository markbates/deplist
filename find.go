package deplist

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobuffalo/envy"
	"github.com/markbates/deplist/internal/oncer"
	"github.com/rogpeppe/go-internal/modfile"
)

// FindImports
func FindImports(dir string, mode build.ImportMode) ([]string, error) {
	if envy.Mods() {
		return viaModules(dir)
	}
	return viaImports(dir, mode)
}

func viaImports(dir string, mode build.ImportMode) ([]string, error) {
	var err error
	var names []string
	cp := envy.CurrentPackage()
	oncer.Do("FindImports"+dir, func() {
		ctx := build.Default

		if len(ctx.SrcDirs()) == 0 {
			err = fmt.Errorf("no src directories found")
			return
		}

		pkg, err := ctx.ImportDir(dir, mode)

		if err != nil {
			if !strings.Contains(err.Error(), "cannot find package") {
				if _, ok := err.(*build.NoGoError); !ok {
					return
				}
			}
		}

		if pkg.ImportPath == "." {
			return
		}
		if pkg.Goroot {
			return
		}
		if len(pkg.GoFiles) <= 0 {
			return
		}

		nm := map[string]string{}
		if !strings.HasPrefix(pkg.ImportPath, cp) {
			nm[pkg.ImportPath] = pkg.ImportPath
		}
		for _, imp := range pkg.Imports {
			if len(ctx.SrcDirs()) == 0 {
				continue
			}
			d := ctx.SrcDirs()[len(ctx.SrcDirs())-1]
			ip := filepath.Join(d, imp)
			n, err := FindImports(ip, mode)
			if err != nil {
				continue
			}
			for _, x := range n {
				nm[x] = x
			}
		}

		var ns []string
		for k := range nm {
			ns = append(ns, k)
		}
		sort.Strings(ns)
		names = ns
	})
	return names, err
}

func viaModules(dir string) ([]string, error) {
	var names []string
	p := filepath.Join(dir, "go.mod")
	moddata, err := ioutil.ReadFile(p)
	if err != nil {
		return names, fmt.Errorf("go.mod cannot be read or does not exist while go module is enabled")
	}
	f, err := modfile.Parse(p, moddata, func(path, version string) (string, error) {
		return version, nil
	})
	if err != nil {
		return names, err
	}
	for _, r := range f.Require {
		if r.Syntax == nil {
			continue
		}
		tok := r.Syntax.Token
		if len(tok) > 1 {
			var ind int
			if tok[0] == "require" {
				ind = 1
			}
			names = append(names, tok[ind])
			continue
		}
	}
	return names, nil
}
