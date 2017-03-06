package deplist

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func List() ([]string, error) {
	deps := []string{}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && filepath.Base(path) != "vendor" {
			cmd := exec.Command("go", "list", "-f", `'* {{ join .Deps  "\n"}}'`, path)
			b, err := cmd.Output()
			if err != nil {
				return err
			}

			list := strings.Split(string(b), "\n")

			for _, g := range list {
				if strings.Contains(g, "github.com") || strings.Contains(g, "bitbucket.org") {
					fmt.Println(g)
					deps = append(deps, g)
				}
			}
		}
		return nil
	})

	sort.Strings(deps)
	return deps, err
}
