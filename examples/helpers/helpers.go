package helpers

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/marco-m/docopt-go"
)

// TestUsage is a helper used to test the output from the examples in this folder.
func TestUsage(usage, command string) {
	opts := docopt.MustParse(usage, strings.Split(command, " ")[1:], "")

	// Sort the keys of the arguments map
	keys := maps.Keys(opts)
	sort.Strings(keys)

	// Print the argument keys and values
	for _, k := range keys {
		fmt.Printf("%9s %v\n", k, opts[k])
	}
	fmt.Println()
}
