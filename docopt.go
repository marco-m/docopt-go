// Licensed under terms of MIT license (see LICENSE-MIT)
// Copyright (c) 2013 Keith Batten, kbatten@gmail.com
// Copyright (c) 2016 David Irvine

package docopt

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Parser struct {
	// OptionsFirst requires that option flags always come before positional
	// arguments; otherwise they can overlap.
	OptionsFirst bool
	// SkipHelpFlags tells the parser not to look for -h and --help flags and
	// call the HelpHandler.
	SkipHelpFlags bool
}

// Parse parses args based on the interface described in doc.
// It never calls os.Exit; you have to handle it yourself, allowing to properly unit
// test command-line arguments. See the examples for idiomatic usage.
// See also: [MustParse].
//
// If you provide a non-empty version string, then this will be displayed when the
// --version flag is found.
func Parse(doc string, args []string, version string) (Opts, error) {
	parser := &Parser{}
	return parser.Parse(doc, args, version)
}

// MustParse parses args based on the interface described in doc.
// If the user asked for help, it prints it and then calls os.Exit(0).
// If the user made an invocation error, it prints the error and calls os.Exit(1).
// See [Parse] for an alternative that allows to unit test and control lifetime..
func MustParse(doc string, argv []string, version string) Opts {
	parser := &Parser{}
	opts, err := parser.Parse(doc, argv, version)
	if err != nil {
		os.Exit(1)
	}
	return opts
}

// Parse parses custom arguments based on the interface described in doc.
// If you provide a non-empty version string, then this will be displayed when
// the --version flag is found.
func (p *Parser) Parse(doc string, argv []string, version string) (Opts, error) {
	return p.parse(doc, argv, version)
}

func (p *Parser) parse(doc string, args []string, version string) (map[string]any, error) {
	opts, output, err := parse(doc, args, !p.SkipHelpFlags, version, p.OptionsFirst)
	var userError *UserError
	if errors.As(err, &userError) {
		// the user gave us bad input
		fmt.Fprintln(os.Stderr, output)
		return opts, err
	}
	// FIXME why are we looking at the len of output? Seems that this information
	//   should instead be encoded only in the error...
	if len(output) > 0 && err == nil {
		// the user asked for help or --version
		fmt.Println(output)
		return opts, ErrHelp
	}
	return opts, err
}

// -----------------------------------------------------------------------------

// parse and return a map of args, output and all errors
func parse(doc string, argv []string, help bool, version string, optionsFirst bool,
) (map[string]any, string, error) {
	if argv == nil && len(os.Args) > 1 {
		argv = os.Args[1:]
	}

	usageSections := parseSection("usage:", doc)

	if len(usageSections) == 0 {
		return nil, "", &LanguageError{`"usage:" (case-insensitive) not found.`}
	}
	if len(usageSections) > 1 {
		return nil, "", &LanguageError{`More than one "usage:" (case-insensitive).`}
	}
	usage := usageSections[0]

	options := parseDefaults(doc)
	formal, err := formalUsage(usage)
	if err != nil {
		return nil, handleError(err, usage), err
	}

	pat, err := parsePattern(formal, &options)
	if err != nil {
		return nil, handleError(err, usage), err
	}

	patternArgv, err := parseArgv(newTokenList(argv, errorUser), &options, optionsFirst)
	if err != nil {
		return nil, handleError(err, usage), err
	}
	patFlat, err := pat.flat(patternOption)
	if err != nil {
		return nil, handleError(err, usage), err
	}
	patternOptions := patFlat.unique()

	patFlat, err = pat.flat(patternOptionSSHORTCUT)
	if err != nil {
		return nil, handleError(err, usage), err
	}
	for _, optionsShortcut := range patFlat {
		docOptions := parseDefaults(doc)
		optionsShortcut.children = docOptions.unique().diff(patternOptions)
	}

	if output := extras(help, version, patternArgv, doc); len(output) > 0 {
		return nil, output, nil
	}

	err = pat.fix()
	if err != nil {
		return nil, handleError(err, usage), err
	}
	matched, left, collected := pat.match(&patternArgv, nil)
	if matched && len(*left) == 0 {
		patFlat, err = pat.flat(patternDefault)
		if err != nil {
			return nil, handleError(err, usage), err
		}
		return append(patFlat, *collected...).dictionary(), "", nil
	}

	// left contains all the non-matched elements, that is, the errors.
	// FIXME
	bho := make([]string, 0, len(*left))
	for _, unknown := range *left {
		if unknown.t == patternOption {
			bho = append(bho, fmt.Sprintf("unknown %s: %s",
				unknown.t, unknown.name))
		} else {
			// FIXME too optimistic ...
			bho = append(bho, fmt.Sprintf("unknown %s: %s %v",
				unknown.t, unknown.name, unknown.value))
		}
	}
	err = &UserError{strings.Join(bho, "\n")}
	return nil, handleError(err, usage), err
}

func handleError(err error, usage string) string {
	if _, ok := err.(*UserError); ok {
		return strings.TrimSpace(fmt.Sprintf("%s\n%s", err, usage))
	}
	return ""
}

func parseSection(name, source string) []string {
	p := regexp.MustCompile(`(?im)^([^\n]*` + name + `[^\n]*\n?(?:[ \t].*?(?:\n|$))*)`)
	s := p.FindAllString(source, -1)
	if s == nil {
		s = []string{}
	}
	for i, v := range s {
		s[i] = strings.TrimSpace(v)
	}
	return s
}

func parseDefaults(doc string) patternList {
	defaults := patternList{}
	p := regexp.MustCompile(`\n[ \t]*(-\S+?)`)
	for _, s := range parseSection("options:", doc) {
		// FIXME corner case "bla: options: --foo"
		_, _, s = stringPartition(s, ":") // get rid of "options:"
		split := p.Split("\n"+s, -1)[1:]
		match := p.FindAllStringSubmatch("\n"+s, -1)
		for i := range split {
			optionDescription := match[i][1] + split[i]
			if strings.HasPrefix(optionDescription, "-") {
				defaults = append(defaults, parseOption(optionDescription))
			}
		}
	}
	return defaults
}

func parsePattern(source string, options *patternList) (*pattern, error) {
	tokens := tokenListFromPattern(source)
	result, err := parseExpr(tokens, options)
	if err != nil {
		return nil, err
	}
	if tokens.current() != nil {
		return nil, tokens.errorFunc("unexpected ending: %s" + strings.Join(tokens.tokens, " "))
	}
	return newRequired(result...), nil
}

func parseArgv(tokens *tokenList, options *patternList, optionsFirst bool) (patternList, error) {
	/*
		Parse command-line argument vector.

		If options_first:
			argv ::= [ long | shorts ]* [ argument ]* [ '--' [ argument ]* ] ;
		else:
			argv ::= [ long | shorts | argument ]* [ '--' [ argument ]* ] ;
	*/
	parsed := patternList{}
	for tokens.current() != nil {
		if tokens.current().eq("--") {
			for _, v := range tokens.tokens {
				parsed = append(parsed, newArgument("", v))
			}
			return parsed, nil
		} else if tokens.current().hasPrefix("--") {
			pl, err := parseLong(tokens, options)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, pl...)
		} else if tokens.current().hasPrefix("-") && !tokens.current().eq("-") {
			ps, err := parseShorts(tokens, options)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, ps...)
		} else if optionsFirst {
			for _, v := range tokens.tokens {
				parsed = append(parsed, newArgument("", v))
			}
			return parsed, nil
		} else {
			parsed = append(parsed, newArgument("", tokens.move().String()))
		}
	}
	return parsed, nil
}

func parseOption(optionDescription string) *pattern {
	optionDescription = strings.TrimSpace(optionDescription)
	options, _, description := stringPartition(optionDescription, "  ")
	options = strings.Replace(options, ",", " ", -1)
	options = strings.Replace(options, "=", " ", -1)

	short := ""
	long := ""
	argcount := 0
	var value any
	value = false

	reDefault := regexp.MustCompile(`(?i)\[default: (.*)\]`)
	for _, s := range strings.Fields(options) {
		if strings.HasPrefix(s, "--") {
			long = s
		} else if strings.HasPrefix(s, "-") {
			short = s
		} else {
			argcount = 1
		}
		if argcount > 0 {
			matched := reDefault.FindAllStringSubmatch(description, -1)
			if len(matched) > 0 {
				value = matched[0][1]
			} else {
				value = nil
			}
		}
	}
	return newOption(short, long, argcount, value)
}

func parseExpr(tokens *tokenList, options *patternList) (patternList, error) {
	// expr ::= seq ( '|' seq )* ;
	seq, err := parseSeq(tokens, options)
	if err != nil {
		return nil, err
	}
	if !tokens.current().eq("|") {
		return seq, nil
	}
	var result patternList
	if len(seq) > 1 {
		result = patternList{newRequired(seq...)}
	} else {
		result = seq
	}
	for tokens.current().eq("|") {
		tokens.move()
		seq, err = parseSeq(tokens, options)
		if err != nil {
			return nil, err
		}
		if len(seq) > 1 {
			result = append(result, newRequired(seq...))
		} else {
			result = append(result, seq...)
		}
	}
	if len(result) > 1 {
		return patternList{newEither(result...)}, nil
	}
	return result, nil
}

func parseSeq(tokens *tokenList, options *patternList) (patternList, error) {
	// seq ::= ( atom [ '...' ] )* ;
	result := patternList{}
	for !tokens.current().match(true, "]", ")", "|") {
		atom, err := parseAtom(tokens, options)
		if err != nil {
			return nil, err
		}
		if tokens.current().eq("...") {
			atom = patternList{newOneOrMore(atom...)}
			tokens.move()
		}
		result = append(result, atom...)
	}
	return result, nil
}

func parseAtom(tokens *tokenList, options *patternList) (patternList, error) {
	// atom ::= '(' expr ')' | '[' expr ']' | 'options' | long | shorts | argument | command ;
	tok := tokens.current()
	result := patternList{}
	if tokens.current().match(false, "(", "[") {
		tokens.move()
		var matching string
		pl, err := parseExpr(tokens, options)
		if err != nil {
			return nil, err
		}
		if tok.eq("(") {
			matching = ")"
			result = patternList{newRequired(pl...)}
		} else if tok.eq("[") {
			matching = "]"
			result = patternList{newOptional(pl...)}
		}
		moved := tokens.move()
		if !moved.eq(matching) {
			return nil, tokens.errorFunc("unmatched '%s', expected: '%s' got: '%s'", tok, matching, moved)
		}
		return result, nil
	} else if tok.eq("options") {
		tokens.move()
		return patternList{newOptionsShortcut()}, nil
	} else if tok.hasPrefix("--") && !tok.eq("--") {
		return parseLong(tokens, options)
	} else if tok.hasPrefix("-") && !tok.eq("-") && !tok.eq("--") {
		return parseShorts(tokens, options)
	} else if tok.hasPrefix("<") && tok.hasSuffix(">") || tok.isUpper() {
		return patternList{newArgument(tokens.move().String(), nil)}, nil
	}
	return patternList{newCommand(tokens.move().String(), false)}, nil
}

func parseLong(tokens *tokenList, options *patternList) (patternList, error) {
	// long ::= '--' chars [ ( ' ' | '=' ) chars ] ;
	long, eq, v := stringPartition(tokens.move().String(), "=")
	var value any
	var opt *pattern
	if eq == "" && v == "" {
		value = nil
	} else {
		value = v
	}

	if !strings.HasPrefix(long, "--") {
		return nil, fmt.Errorf("long option '%s' doesn't start with --", long)
	}
	similar := patternList{}
	for _, o := range *options {
		if o.long == long {
			similar = append(similar, o)
		}
	}
	if tokens.err == errorUser && len(similar) == 0 { // if no exact match
		similar = patternList{}
		for _, o := range *options {
			if strings.HasPrefix(o.long, long) {
				similar = append(similar, o)
			}
		}
	}
	if len(similar) > 1 { // might be simply specified ambiguously 2+ times?
		similarLong := make([]string, len(similar))
		for i, s := range similar {
			similarLong[i] = s.long
		}
		return nil, tokens.errorFunc("%s is not a unique prefix: %s?", long, strings.Join(similarLong, ", "))
	} else if len(similar) < 1 {
		argcount := 0
		if eq == "=" {
			argcount = 1
		}
		opt = newOption("", long, argcount, false)
		*options = append(*options, opt)
		if tokens.err == errorUser {
			var val any
			if argcount > 0 {
				val = value
			} else {
				val = true
			}
			opt = newOption("", long, argcount, val)
		}
	} else {
		opt = newOption(similar[0].short, similar[0].long, similar[0].argcount, similar[0].value)
		if opt.argcount == 0 {
			if value != nil {
				return nil, tokens.errorFunc("%s must not have an argument", opt.long)
			}
		} else {
			if value == nil {
				if tokens.current().match(true, "--") {
					return nil, tokens.errorFunc("%s requires argument", opt.long)
				}
				moved := tokens.move()
				if moved != nil {
					value = moved.String() // only set as string if not nil
				}
			}
		}
		if tokens.err == errorUser {
			if value != nil {
				opt.value = value
			} else {
				opt.value = true
			}
		}
	}

	return patternList{opt}, nil
}

func parseShorts(tokens *tokenList, options *patternList) (patternList, error) {
	// shorts ::= '-' ( chars )* [ [ ' ' ] chars ] ;
	tok := tokens.move()
	if !tok.hasPrefix("-") || tok.hasPrefix("--") {
		return nil, fmt.Errorf("short option '%s' doesn't start with -", tok)
	}
	left := strings.TrimLeft(tok.String(), "-")
	parsed := patternList{}
	for left != "" {
		var opt *pattern
		short := "-" + left[0:1]
		left = left[1:]
		similar := patternList{}
		for _, o := range *options {
			if o.short == short {
				similar = append(similar, o)
			}
		}
		if len(similar) > 1 {
			return nil, tokens.errorFunc("%s is specified ambiguously %d times", short, len(similar))
		} else if len(similar) < 1 {
			opt = newOption(short, "", 0, false)
			*options = append(*options, opt)
			if tokens.err == errorUser {
				opt = newOption(short, "", 0, true)
			}
		} else { // why copying is necessary here?
			opt = newOption(short, similar[0].long, similar[0].argcount, similar[0].value)
			var value any
			if opt.argcount > 0 {
				if left == "" {
					if tokens.current().match(true, "--") {
						return nil, tokens.errorFunc("%s requires argument", short)
					}
					value = tokens.move().String()
				} else {
					value = left
					left = ""
				}
			}
			if tokens.err == errorUser {
				if value != nil {
					opt.value = value
				} else {
					opt.value = true
				}
			}
		}
		parsed = append(parsed, opt)
	}
	return parsed, nil
}

func formalUsage(section string) (string, error) {
	_, _, section = stringPartition(section, ":") // drop "usage:"
	pu := strings.Fields(section)

	if len(pu) == 0 {
		return "", &LanguageError{"no fields found in usage (perhaps a spacing error)."}
	}

	result := "( "
	for _, s := range pu[1:] {
		if s == pu[0] {
			result += ") | ( "
		} else {
			result += s + " "
		}
	}
	result += ")"

	return result, nil
}

func extras(help bool, version string, options patternList, doc string) string {
	if help {
		for _, o := range options {
			if (o.name == "-h" || o.name == "--help") && o.value == true {
				return strings.Trim(doc, "\n")
			}
		}
	}
	if version != "" {
		for _, o := range options {
			if (o.name == "--version") && o.value == true {
				return version
			}
		}
	}
	return ""
}

func stringPartition(s, sep string) (string, string, string) {
	sepPos := strings.Index(s, sep)
	if sepPos == -1 { // no separator found
		return s, "", ""
	}
	split := strings.SplitN(s, sep, 2)
	return split[0], sep, split[1]
}
