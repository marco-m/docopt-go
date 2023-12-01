package docopt

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

func ExampleMustParse() {
	usage := `
Usage:
  example tcp [<host>...] [--force] [--timeout=<seconds>]
  example serial <port> [--baud=<rate>] [--timeout=<seconds>]
  example --help | --version`

	// Parse the command line `example tcp 127.0.0.1 --force`
	argv := []string{"tcp", "127.0.0.1", "--force"}
	opts := MustParse(usage, argv, "0.1.1rc")

	// Sort the keys of the options map
	keys := maps.Keys(opts)
	sort.Strings(keys)

	// Print the option keys and values
	for _, k := range keys {
		fmt.Printf("%9s %v\n", k, opts[k])
	}

	// Output:
	//    --baud <nil>
	//   --force true
	//    --help false
	// --timeout <nil>
	// --version false
	//    <host> [127.0.0.1]
	//    <port> <nil>
	//    serial false
	//       tcp true
}

func ExampleOpts_Bind() {
	usage := `
Usage:
  example tcp [<host>...] [--force] [--timeout=<seconds>]
  example serial <port> [--baud=<rate>] [--timeout=<seconds>]
  example --help | --version`

	// Parse the command line `example serial 443 --baud=9600`
	argv := []string{"serial", "443", "--baud=9600"}
	opts, err := Parse(usage, argv, "0.1.1rc")
	if err != nil {
		// In real code you would not panic; we do this here due to how
		// Go testable examples work.
		panic(err)
	}

	var conf struct {
		Tcp     bool
		Serial  bool
		Host    []string
		Port    int
		Force   bool
		Timeout int
		Baud    int
	}
	if err := opts.Bind(&conf); err != nil {
		// In real code you would not panic; we do this here due to how
		// Go testable examples work.
		panic(err)
	}

	if conf.Serial {
		fmt.Printf("port: %d, baud: %d", conf.Port, conf.Baud)
	}

	// Output:
	// port: 443, baud: 9600
}
