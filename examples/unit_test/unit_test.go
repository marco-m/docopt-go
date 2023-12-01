package main

import (
	"reflect"
	"testing"

	"github.com/marco-m/docopt-go"
)

const usage = `Usage:
  nettool tcp <host> <port> [--timeout=<seconds>]
  nettool serial <port> [--baud=9600] [--timeout=<seconds>]
  nettool -h | --help | --version`

// This is an example of how to write unit tests for a client of docopt.
// As such, we use only the stdlib testing package.
//
// In Go, it is better to have separate test functions for the success and
// failure cases; do not try to cram everything together.
func TestUsageSuccess(t *testing.T) {
	type testCase struct {
		name string
		argv []string
		want docopt.Opts
	}

	test := func(t *testing.T, tc testCase) {
		opts, err := docopt.Parse(usage, tc.argv, "")

		if err != nil {
			t.Fatalf("have: %s; want: <no error>", err)
		}
		if !reflect.DeepEqual(opts, tc.want) {
			t.Fatalf("have: %v; want: %v", opts, tc.want)
		}
	}

	testCases := []testCase{
		{
			name: "tcp",
			argv: []string{"tcp", "myhost.com", "8080", "--timeout=20"},
			want: docopt.Opts{
				"--baud":    nil,
				"--help":    false,
				"--timeout": "20",
				"--version": false,
				"-h":        false,
				"<host>":    "myhost.com",
				"<port>":    "8080",
				"serial":    false,
				"tcp":       true,
			},
		},
		{
			name: "serial",
			argv: []string{"serial", "1234", "--baud=14400"},
			want: docopt.Opts{
				"--baud":    "14400",
				"--help":    false,
				"--timeout": nil,
				"--version": false,
				"-h":        false,
				"<host>":    nil,
				"<port>":    "1234",
				"serial":    true,
				"tcp":       false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

// This is an example of how to write unit tests for a client of docopt.
// As such, we use only the stdlib testing package.
//
// In Go, it is better to have separate test functions for the success and
// failure cases; do not try to cram everything together.
func TestUsageFailure(t *testing.T) {
	type testCase struct {
		name string
		argv []string
		// TODO should also discriminate on the error type
		// wantErr error
	}

	test := func(t *testing.T, tc testCase) {
		_, err := docopt.Parse(usage, tc.argv, "")

		if err == nil {
			t.Fatalf("have: <no error>; want: some error")
		}
	}

	testCases := []testCase{
		{
			name: "foo",
			argv: []string{"foo", "bar", "dog"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
