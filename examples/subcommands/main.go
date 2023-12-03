// This is an example of how to write unit tests for a client of docopt
// that uses doctopt.Parse

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/marco-m/docopt-go"
)

const usage = `
nettool connects to a remote host via various protocols.

Usage:
  nettool [options] <command> [<args>...]
  nettool -h | --help | --version

Options:
  --timeout=<secs>  Connection timeout in seconds [default: 10].

Commands:
  tcp     Connect via TCP.
  serial  Connect via serial.
`

func main() {
	os.Exit(mainInt())
}

func mainInt() int {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func run(args []string) error {
	var cfg struct {
		Command string
		Args    []string
		Timeout int
	}

	// OptionsFirst is needed to make the `-h` work correctly for subcommands :-/
	parser := docopt.Parser{OptionsFirst: true}
	opts, err := parser.Parse(usage, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	// workaround: remove global options from the command line
	subArgs := append([]string{cfg.Command}, cfg.Args...)
	switch cfg.Command {
	case "tcp":
		return cmdTcp(subArgs, cfg.Timeout)
	case "serial":
		return cmdSerial(subArgs, cfg.Timeout)
	default:
		return fmt.Errorf("unknown command: %s", cfg.Command)
	}
}

const usageTcp = `
Usage:
  nettool tcp <host> <port>
`

func cmdTcp(args []string, timeout int) error {
	var cfg struct {
		Tcp  bool // sigh workaround should not be required
		Host string
		Port int
	}

	opts, err := docopt.Parse(usageTcp, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	fmt.Printf("Connecting to %s:%d via TCP with timeout=%d ...\n",
		cfg.Host, cfg.Port, timeout)

	return nil
}

const usageSerial = `
Usage:
  nettool serial [options] <port>

Options:
  --baud=<speed>  Connection speed in bauds [default: 9600].
`

func cmdSerial(args []string, timeout int) error {
	var cfg struct {
		Serial bool // sigh workaround should not be required
		Port   int
		Baud   int
	}

	opts, err := docopt.Parse(usageSerial, args, "")
	if errors.Is(err, docopt.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = opts.Bind(&cfg); err != nil {
		return err
	}

	fmt.Printf("Connecting to :%d via serial with baud=%d timeout=%d  ...\n",
		cfg.Port, cfg.Baud, timeout)

	return nil
}
