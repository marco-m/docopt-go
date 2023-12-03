package docopt

import "errors"

// ErrHelp can be used in client code to decide when to call os.Exit(0) as opposed
// to os.Exit(1). For example:
//
//		opts, err := docopt.Parse(usage, os.Args[1:], "")
//		if err != nil {
//			if errors.Is(err, docopt.ErrHelp) {
//				os.Exit(0)
//			}
//	     fmt.FPrintln(os.Stderr, err)
//			os.Exit(1)
//		}
//	 ...
var ErrHelp = errors.New("user requested help")

// UserError records an error with program arguments.
// Can be used also by client code to report specific CLI validation errors.
type UserError struct {
	Msg string
}

func (e UserError) Error() string {
	return e.Msg
}

// ErrLanguage represents an error with the doc string (the program is wrong).
var ErrLanguage = errors.New("")
