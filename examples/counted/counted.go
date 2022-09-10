package main

import (
	"fmt"
	"os"

	"github.com/marco-m/docopt-go"
)

var usage = `Usage: counted --help
       counted -v...
       counted go [go]
       counted (--path=<path>)...
       counted <file> <file>

Try: counted -vvvvvvvvvv
     counted go go
     counted --path ./here --path ./there
     counted this.txt that.txt`

func main() {
	arguments := docopt.MustParse(usage, os.Args[1:], "")
	fmt.Println(arguments)
}
