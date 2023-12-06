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
		fmt.Fprintln(os.Stderr, err)
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
	usage, err := extractUsage(doc)
	if err != nil {
		return nil, err
	}

	opts, err := parse(doc, args, !p.SkipHelpFlags, version, p.OptionsFirst)

	if errors.Is(err, ErrUser) {
		// the user gave us bad input
		//fmt.Fprintln(os.Stderr, err)
		//fmt.Fprintln(os.Stderr, usage)
		//return nil, err
		return nil, fmt.Errorf("%w\n%s", err, usage)
	}

	if errors.Is(err, ErrHelp) {
		fmt.Println(err)
		return nil, err
	}

	return opts, err
}

// -----------------------------------------------------------------------------

func extractUsage(doc string) (string, error) {
	usageSections := parseSection("usage:", doc)
	if len(usageSections) == 0 {
		return "", fmt.Errorf("%s%w", "section 'usage' not found", ErrLanguage)
	}
	if len(usageSections) > 1 {
		return "", fmt.Errorf("%s%w", "more than one section 'usage'", ErrLanguage)
	}
	return usageSections[0], nil
}

// parse and return a map of args, output and all errors
func parse(doc string, argv []string, help bool, version string, optionsFirst bool,
) (map[string]any, error) {
	if argv == nil {
		return nil, fmt.Errorf("%s%w", "nil command-line", ErrLanguage)
	}
	usage, err := extractUsage(doc)
	if err != nil {
		return nil, err
	}
	formal, err := formalUsage(usage)
	if err != nil {
		return nil, err
	}

	options := parseDefaults(doc)
	pat, err := parsePattern(formal, &options)
	if err != nil {
		return nil, err
	}

	patternArgv, err := parseArgv(newTokenList(argv, errorTypeUser), &options, optionsFirst)
	if err != nil {
		return nil, err
	}
	patFlat, err := pat.flat(patternOption)
	if err != nil {
		return nil, err
	}
	patternOptions := patFlat.unique()

	patFlat, err = pat.flat(patternOptionSSHORTCUT)
	if err != nil {
		return nil, err
	}
	for _, optionsShortcut := range patFlat {
		docOptions := parseDefaults(doc)
		optionsShortcut.Children = docOptions.unique().diff(patternOptions)
	}

	if output := extras(help, version, patternArgv, doc); len(output) > 0 {
		return nil, fmt.Errorf("%s%w", output, ErrHelp)
	}

	err = pat.fix()
	if err != nil {
		return nil, err
	}
	matched, left, collected := pat.match(&patternArgv, nil)
	if matched && len(*left) == 0 {
		patFlat, err = pat.flat(patternDefault)
		if err != nil {
			return nil, err
		}
		return append(patFlat, *collected...).dictionary(), nil
	}

	// left contains all the non-matched elements, that is, the errors.
	// FIXME
	bho := make([]string, 0, len(*left))
	for _, unknown := range *left {
		if unknown.Type == patternOption {
			bho = append(bho, fmt.Sprintf("unknown %s: %s",
				unknown.Type, unknown.Name))
		} else {
			// FIXME too optimistic ...
			bho = append(bho, fmt.Sprintf("unknown %s: %s %v",
				unknown.Type, unknown.Name, unknown.Value))
		}
	}
	err = fmt.Errorf("%s%w", strings.Join(bho, "\n"), ErrUser)
	return nil, err
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
	options = strings.ReplaceAll(options, ",", " ")
	options = strings.ReplaceAll(options, "=", " ")

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
		if o.Long == long {
			similar = append(similar, o)
		}
	}
	if tokens.err == errorTypeUser && len(similar) == 0 { // if no exact match
		similar = patternList{}
		for _, o := range *options {
			if strings.HasPrefix(o.Long, long) {
				similar = append(similar, o)
			}
		}
	}
	if len(similar) > 1 { // might be simply specified ambiguously 2+ times?
		similarLong := make([]string, len(similar))
		for i, s := range similar {
			similarLong[i] = s.Long
		}
		return nil, tokens.errorFunc("%s is not a unique prefix: %s?", long, strings.Join(similarLong, ", "))
	} else if len(similar) < 1 {
		argcount := 0
		if eq == "=" {
			argcount = 1
		}
		opt = newOption("", long, argcount, false)
		*options = append(*options, opt)
		if tokens.err == errorTypeUser {
			var val any
			if argcount > 0 {
				val = value
			} else {
				val = true
			}
			opt = newOption("", long, argcount, val)
		}
	} else {
		opt = newOption(similar[0].Short, similar[0].Long, similar[0].ArgCount, similar[0].Value)
		if opt.ArgCount == 0 {
			if value != nil {
				return nil, tokens.errorFunc("%s must not have an argument", opt.Long)
			}
		} else {
			if value == nil {
				if tokens.current().match(true, "--") {
					return nil, tokens.errorFunc("%s requires argument", opt.Long)
				}
				moved := tokens.move()
				if moved != nil {
					value = moved.String() // only set as string if not nil
				}
			}
		}
		if tokens.err == errorTypeUser {
			if value != nil {
				opt.Value = value
			} else {
				opt.Value = true
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
			if o.Short == short {
				similar = append(similar, o)
			}
		}
		if len(similar) > 1 {
			return nil, tokens.errorFunc("%s is specified ambiguously %d times", short, len(similar))
		} else if len(similar) < 1 {
			opt = newOption(short, "", 0, false)
			*options = append(*options, opt)
			if tokens.err == errorTypeUser {
				opt = newOption(short, "", 0, true)
			}
		} else { // why copying is necessary here?
			opt = newOption(short, similar[0].Long, similar[0].ArgCount, similar[0].Value)
			var value any
			if opt.ArgCount > 0 {
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
			if tokens.err == errorTypeUser {
				if value != nil {
					opt.Value = value
				} else {
					opt.Value = true
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
		// FIXME find better error message
		return "", fmt.Errorf("%s%w",
			"no fields found in section 'usage' (perhaps a spacing error)", ErrLanguage)
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
			if (o.Name == "-h" || o.Name == "--help") && o.Value == true {
				return strings.Trim(doc, "\n")
			}
		}
	}
	if version != "" {
		for _, o := range options {
			if (o.Name == "--version") && o.Value == true {
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
