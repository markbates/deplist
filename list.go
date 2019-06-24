package deplist

import (
	"fmt"
	"go/build"
	"os"

	"github.com/markbates/deplist/internal/takeon/github.com/markbates/oncer"
)

func List(skips ...string) (map[string]string, error) {
	oncer.Do("deplist.List", func() {
		fmt.Println("deplist.List has been deprecated. Use deplist.FindImports instead")
	})
	deps := map[string]string{}
	pwd, err := os.Getwd()
	if err != nil {
		return deps, err
	}
	pkgs, err := FindImports(pwd, build.IgnoreVendor)
	if err != nil {
		return deps, err
	}
	for _, p := range pkgs {
		deps[p] = p
	}

	return deps, nil
}
