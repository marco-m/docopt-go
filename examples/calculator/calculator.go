package main

import (
	"fmt"
	"os"

	"github.com/marco-m/docopt-go"
)

var usage = `Not a serious example.

Usage:
  calculator <value> ( ( + | - | * | / ) <value> )...
  calculator <function> <value> [( , <value> )]...
  calculator (-h | --help)

Examples:
  calculator 1 + 2 + 3 + 4 + 5
  calculator 1 + 2 '*' 3 / 4 - 5    # note quotes around '*'
  calculator sum 10 , 20 , 30 , 40

Options:
  -h, --help
`

func main() {
	arguments := docopt.MustParse(usage, os.Args[1:], "")
	fmt.Println(arguments)
}
