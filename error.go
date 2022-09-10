package docopt

import (
	"fmt"
)

// UserError records an error with program arguments.
// Can be used by client code to perform specific CLI validation.
type UserError struct {
	Msg string
}

func (e UserError) Error() string {
	return e.Msg
}

// LanguageError records an error with the doc string.
type LanguageError struct {
	msg string
}

func (e LanguageError) Error() string {
	return e.msg
}

var newError = fmt.Errorf
