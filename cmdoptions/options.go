package cmdoptions

import (
	"errors"
	"flag"
	"io"
	"os"
	"time"
)

// InputOptions holds input arguments.
type InputOptions struct {
	// TestsuiteFile is path of the testsuite JSON file.
	TestsuiteFile string
	// Workers represents number of concurrent workers.
	Workers int
	// Timeout is the maximum duration for each HTTP request.
	Timeout time.Duration
	// StopOnFailure force exits on first error or failure.
	StopOnFailure bool
}

var usage = `
Usage: smoker [options...]

Example:
  smoker -testsuite smoketestsuite-api.json
  smoker -testsuite smoketestsuite-web.json -workers 4 -timeout 5 -stop-on-failure

Options:
  -testsuite        Testsuite file in JSON format to read test cases.
  -workers          Number of workers to send requests concurrently. (accepts integer value >= 1. Default is 1. 0 is not allowed)
  -timeout          Set timeout per request in seconds. (accepts integer value >= 1. Default is 10. 0 is not allowed)
  -stop-on-failure  Stop execution upon first error or failure.
  -version          Prints the version and exits.

Visit: https://github.com/amad/smoker
`

var versionFlag bool

// InstallFlags adds CLI flags and validates user input.
func InstallFlags(version string, stdout io.StringWriter) (*InputOptions, error) {
	var flags InputOptions

	addOptions(&flags, stdout)

	if versionFlag {
		stdout.WriteString(version + "\n")
		os.Exit(0)
	}

	if flags.TestsuiteFile == "" {
		return &flags, errors.New("-testsuite is required")
	}

	if flags.Workers < 1 {
		return &flags, errors.New("-workers only accept a number >= 1")
	}

	if flags.Timeout < 1 {
		return &flags, errors.New("-timeout only accept a number >= 1")
	}

	return &flags, nil
}

func addOptions(flags *InputOptions, stdout io.StringWriter) {
	var timeout int

	flag.IntVar(&flags.Workers, "workers", 1, "")
	flag.IntVar(&timeout, "timeout", 10, "")
	flag.BoolVar(&versionFlag, "version", false, "")
	flag.StringVar(&flags.TestsuiteFile, "testsuite", "", "")
	flag.BoolVar(&flags.StopOnFailure, "stop-on-failure", false, "")

	flag.Usage = func() {
		stdout.WriteString(usage)
		os.Exit(0)
	}

	flag.Parse()

	flags.Timeout = time.Duration(timeout) * time.Second
}
