package docopt_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/docopt-go"
)

const topCommandHelp = `
Usage:
  shop [options] [<args>...]
  shop -h

Options:
  --organic  Buy the organic version.`

const topCommandUsage = `
Usage:
  shop [options] [<args>...]
  shop -h`

func TestParseTopCommandSuccess(t *testing.T) {
	type testCase struct {
		name string
		args []string
		want docopt.Opts
	}

	test := func(t *testing.T, tc testCase) {
		opts, err := docopt.Parse(topCommandHelp, tc.args, "")

		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.DeepEquals(opts, tc.want))
	}

	testCases := []testCase{
		{
			name: "no objects",
			args: []string{""},
			want: docopt.Opts{
				"--organic": false,
				"-h":        false,
				"<args>":    []string{""},
			},
		},
		{
			name: "one object",
			args: []string{"banana"},
			want: docopt.Opts{
				"--organic": false,
				"-h":        false,
				"<args>":    []string{"banana"},
			},
		},
		{
			name: "two object",
			args: []string{"banana", "mango"},
			want: docopt.Opts{
				"--organic": false,
				"-h":        false,
				"<args>":    []string{"banana", "mango"},
			},
		},
		{
			name: "one object with flag",
			args: []string{"--organic", "banana"},
			want: docopt.Opts{
				"--organic": true,
				"-h":        false,
				"<args>":    []string{"banana"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestParseTopCommandFailure(t *testing.T) {
	type testCase struct {
		name       string
		args       []string
		wantErr    error
		wantErrMsg string
	}

	test := func(t *testing.T, tc testCase) {
		_, err := docopt.Parse(topCommandHelp, tc.args, "")

		qt.Assert(t, qt.ErrorIs(err, tc.wantErr))
		qt.Assert(t, qt.Equals(err.Error(), tc.wantErrMsg))
	}

	testCases := []testCase{
		{
			name:       "ask for help",
			args:       []string{"-h"},
			wantErr:    docopt.ErrHelp,
			wantErrMsg: strings.TrimSpace(topCommandHelp),
		},
		{
			name:       "unknown option",
			args:       []string{"--foo"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown option: --foo" + topCommandUsage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

const subCommandHelp = `
nettool connects to a remote host via various protocols.

Usage:
  nettool [options] <command> [<args>...]
  nettool -h

Options:
  --timeout=<secs>  Connection timeout in seconds [default: 10].

Commands:
  tcp     Connect via TCP.
  serial  Connect via serial.`

const subCommandUsage = `
Usage:
  nettool [options] <command> [<args>...]
  nettool -h`

const tcpHelp = `
Usage:
  nettool tcp <host> <port>`

const tcpUsage = `
Usage:
  nettool tcp <host> <port>`

const serialHelp = `
Usage:
  nettool serial [options] <port>

Options:
  --baud=<speed>  Connection speed in bauds [default: 9600].`

const serialUsage = `
Usage:
  nettool serial [options] <port>`

func subCommandRun(args []string) error {
	var cfg struct {
		Command string
		Args    []string
		Timeout int
	}

	// OptionsFirst is needed to make the `-h` work correctly for subcommands :-/
	parser := docopt.Parser{OptionsFirst: true}
	opts, err := parser.Parse(subCommandHelp, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		// In prod we would return nil, but in this case we return the error
		// to allow the test to compare it.
		return err
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	// workaround: remove global options from the command line
	subArgs := append([]string{cfg.Command}, cfg.Args...)
	// FIXME this will fix the 2 and break others...
	//subArgs := cfg.Args
	switch cfg.Command {
	case "tcp":
		return cmdTcp(subArgs, cfg.Timeout)
	case "serial":
		return cmdSerial(subArgs, cfg.Timeout)
	case "":
		return fmt.Errorf("missing command%w", docopt.ErrUser)
	default:
		return fmt.Errorf("unknown command: %s%w", cfg.Command, docopt.ErrUser)
	}
}

func cmdTcp(args []string, timeout int) error {
	var cfg struct {
		Tcp  bool // sigh workaround should not be required
		Host string
		Port int
	}

	opts, err := docopt.Parse(tcpHelp, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		// In prod we would return nil, but in this case we return the error
		// to allow the test to compare it.
		return err
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	// do something here
	return nil
}

func cmdSerial(args []string, timeout int) error {
	var cfg struct {
		Serial bool // sigh workaround should not be required
		Port   int
		Baud   int
	}

	opts, err := docopt.Parse(serialHelp, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		// In prod we would return nil, but in this case we return the error
		// to allow the test to compare it.
		return err
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	// do something here
	return nil
}

func TestParseSubCommandSuccess(t *testing.T) {
	type testCase struct {
		name string
		args []string
		want docopt.Opts
	}

	test := func(t *testing.T, tc testCase) {
		err := subCommandRun(tc.args)

		qt.Assert(t, qt.IsNil(err))
	}

	testCases := []testCase{
		{
			name: "no arguments",
			args: []string{""},
			want: docopt.Opts{
				"--help":    false,
				"--timeout": "10",
				"--version": false,
				"-h":        false,
				"<args>":    []string{},
				"<command>": "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestParseSubCommandFailure(t *testing.T) {
	type testCase struct {
		name       string
		args       []string
		wantErr    error
		wantErrMsg string
	}

	test := func(t *testing.T, tc testCase) {
		err := subCommandRun(tc.args)

		qt.Assert(t, qt.ErrorIs(err, tc.wantErr))
		qt.Assert(t, qt.Equals(err.Error(), tc.wantErrMsg))
	}

	testCases := []testCase{
		{
			name:       "missing command top level",
			args:       []string{""},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "missing command",
		},
		{
			name:       "unknown command top level",
			args:       []string{"foo"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown command: foo",
		},
		{
			name:       "unknown option top level",
			args:       []string{"--foo"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown option: --foo" + subCommandUsage,
		},
		{
			name:       "help requested top level",
			args:       []string{"-h"},
			wantErr:    docopt.ErrHelp,
			wantErrMsg: strings.TrimSpace(subCommandHelp),
		},
		{
			name:       "help requested for tcp",
			args:       []string{"tcp", "-h"},
			wantErr:    docopt.ErrHelp,
			wantErrMsg: strings.TrimSpace(tcpHelp),
		},
		{
			name:       "help requested for serial",
			args:       []string{"serial", "-h"},
			wantErr:    docopt.ErrHelp,
			wantErrMsg: strings.TrimSpace(serialHelp),
		},
		{
			name:       "too few arguments for tcp (0)",
			args:       []string{"tcp"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "XXX" + tcpUsage, // FIXME BROKEN
		},
		{
			name:       "too few arguments for tcp (1)",
			args:       []string{"tcp", "host"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "XXX" + tcpUsage, // FIXME BROKEN
		},
		{
			name:       "too few arguments for serial",
			args:       []string{"serial"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "XXX" + serialUsage, // FIXME BROKEN
		},
		{
			name:       "unknown option for tcp",
			args:       []string{"tcp", "--foo"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown option: --foo" + tcpUsage, // FIXME BROKEN
		},
		{
			name:       "unknown option for serial",
			args:       []string{"serial", "--foo"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown option: --foo" + serialUsage, // FIXME BROKEN
		},
		{
			name:       "too many arguments for tcp",
			args:       []string{"tcp", "host", "123", "too-much"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown argument: too-much" + tcpUsage,
		},
		{
			name:       "too many arguments for serial",
			args:       []string{"serial", "123", "too-much"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown argument: too-much" + serialUsage,
		},
		{
			name:       "too many arguments and unknown option for serial",
			args:       []string{"serial", "--foo", "123", "too-much"},
			wantErr:    docopt.ErrUser,
			wantErrMsg: "unknown option: --foo\nunknown argument: too-much" + serialUsage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
