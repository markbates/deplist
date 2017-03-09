package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/markbates/deplist"
)

func main() {
	deps, err := deplist.List()
	if err != nil {
		log.Fatal(err)
	}
	list := make([]string, 0, len(deps))
	for k := range deps {
		list = append(list, k)
	}
	sort.Strings(list)
	fmt.Println(strings.Join(list, "\n"))
}
