package main

import (
	"fmt"

	"github.com/marco-m/docopt-go"
)

func main() {
	usage := `Usage:
  quick tcp <host> <port> [--timeout=<seconds>]
  quick serial <port> [--baud=9600] [--timeout=<seconds>]
  quick -h | --help | --version`

	arguments := docopt.MustParse(usage, nil, "0.1.1rc")
	fmt.Println(arguments)
}
