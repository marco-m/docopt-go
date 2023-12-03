package docopt

import "errors"

// ErrHelp can be used in client code to decide when to call os.Exit(0) as opposed
// to os.Exit(1). For example:
//
//	opts, err := docopt.Parse(usage, os.Args[1:], "")
//	if errors.Is(err, docopt.ErrHelp) {
//	    os.Exit(0)
//	}
//	if err != nil {
//	    fmt.FPrintln(os.Stderr, err)
//	    os.Exit(1)
//	}
//	...
var ErrHelp = errors.New("")

// ErrUser represents an error in the command-line arguments (the user made an
// invocation error).
//
// If desired, client code can report specific CLI validation errors by
// wrapping ErrUser as follows:
//
//	fmt.Errorf("%s%w", "write specific error here", ErrUser)
//
// Note: do not add spaces around the wrapping verb %w.
var ErrUser = errors.New("")

// ErrLanguage represents an error with the doc string (the program is wrong).
var ErrLanguage = errors.New("")
