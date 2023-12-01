// This file runs tests using the 'testscript' package.
// To understand, see:
// - https://github.com/rogpeppe/go-internal
// - https://bitfieldconsulting.com/golang/test-scripts

package docopt_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/marco-m/docopt-go"
)

func TestMain(m *testing.M) {
	// The commands map holds the set of command names, each with an associated
	// run function which should return the code to pass to os.Exit.
	// When [testscript.Run] is called, these commands are installed as regular
	// commands in the shell path, so can be invoked with "exec".
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"mustparse": mustparseMain,
	}))
}

func TestScriptDocopt(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func mustparseMain() int {
	usage := `
Usage:
  mustparse tcp [<host>...] [--timeout=<seconds>]`

	opts := docopt.MustParse(usage, os.Args[1:], "")
	fmt.Println("SUCCESS!")
	fmt.Println(opts)
	return 0
}
