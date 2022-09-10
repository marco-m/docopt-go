package main

import (
	"fmt"
	"os"

	"github.com/marco-m/docopt-go"
)

func main() {
	usage := `Usage: odd_even [-h | --help] (ODD EVEN)...

Example, try:
  odd_even 1 2 3 4

Options:
  -h, --help`

	arguments, _ := docopt.Parse(usage, os.Args[1:], "")
	fmt.Println(arguments)
}
