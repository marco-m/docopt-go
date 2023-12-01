package main

import (
	"github.com/marco-m/docopt-go/examples/helpers"
)

func Example() {
	helpers.TestUsage(usage, "counted -vvvvvvvvvv")
	helpers.TestUsage(usage, "counted go go")
	helpers.TestUsage(usage, "counted --path ./here --path ./there")
	helpers.TestUsage(usage, "counted this.txt that.txt")
	// Output:
	//    --help false
	//    --path []
	//        -v 10
	//    <file> []
	//        go 0
	//
	//    --help false
	//    --path []
	//        -v 0
	//    <file> []
	//        go 2
	//
	//    --help false
	//    --path [./here ./there]
	//        -v 0
	//    <file> []
	//        go 0
	//
	//    --help false
	//    --path []
	//        -v 0
	//    <file> [this.txt that.txt]
	//        go 0
}
