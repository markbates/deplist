package deplist

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var commonSkips = []string{"vendor", ".git", "examples", "node_modules"}

func List(skips ...string) (map[string]string, error) {
	wg := &sync.WaitGroup{}
	moot := &sync.Mutex{}

	skips = append(skips, commonSkips...)
	deps := map[string]string{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		for _, s := range skips {
			if strings.Contains(path, s) {
				return filepath.SkipDir
			}
		}
		if info.IsDir() {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				cmd := exec.Command("go", "list", "-f", `'* {{ join .Deps  "\n"}}'`, "./"+path)
				// fmt.Println(strings.Join(cmd.Args, " "))
				b, err := cmd.Output()
				if err != nil {
					return
				}

				list := strings.Split(string(b), "\n")

				for _, g := range list {
					if strings.Contains(g, "github.com") || strings.Contains(g, "bitbucket.org") {
						g = strings.TrimPrefix(g, "'* ")
						moot.Lock()
						skip := false
						for _, s := range skips {
							if strings.Contains(g, s) {
								skip = true
								break
							}
						}
						if !skip {
							deps[g] = g
						}
						moot.Unlock()
					}
				}
			}(path)
		}
		return nil
	})

	wg.Wait()

	return deps, err
}
