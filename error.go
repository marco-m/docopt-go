package docopt

// UserError records an error with program arguments.
// Can be used also by client code to report specific CLI validation errors.
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
