package main

import "github.com/marco-m/docopt-go/examples/helpers"

func Example() {
	helpers.TestUsage(usage, "arguments -qv")
	helpers.TestUsage(usage, "arguments --left file.A file.B")
	// Output:
	//    --help false
	//    --left false
	//   --right false
	//        -q true
	//        -r false
	//        -v true
	// CORRECTION <nil>
	//      FILE []
	//
	//    --help false
	//    --left true
	//   --right false
	//        -q false
	//        -r false
	//        -v false
	// CORRECTION file.A
	//      FILE [file.B]
}
