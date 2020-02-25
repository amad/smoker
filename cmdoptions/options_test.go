package cmdoptions

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestInstallFlags(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name      string
		args      []string
		options   *InputOptions
		expectErr string
	}{
		{"flagset1", []string{"app", "-testsuite", "test"}, &InputOptions{"test", 1, time.Duration(10) * time.Second, false}, ""},
		{"flagset2", []string{"app", "-testsuite", "test", "-workers", "2", "-timeout", "5", "-stop-on-failure"}, &InputOptions{"test", 2, time.Duration(5) * time.Second, true}, ""},
		{"no_args", []string{"app", ""}, nil, "-testsuite is required"},
		{"no_testsuite", []string{"app", "-stop-on-failure", "0"}, nil, "-testsuite is required"},
		{"invalid_workers", []string{"app", "-workers", "0", "-testsuite", "test"}, nil, "-workers only accept a number >= 1"},
		{"invalid_timeout", []string{"app", "-timeout", "0", "-testsuite", "test"}, nil, "-timeout only accept a number >= 1"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			var buffer bytes.Buffer

			originalArgs := os.Args

			os.Args = tc.args
			options, err := InstallFlags("version", &buffer)

			os.Args = originalArgs

			if err != nil {
				if tc.expectErr == "" {
					t.Fatalf("Unexpected error\nexpected: <nil>\nreceived: %s", err.Error())
				}

				if tc.expectErr != err.Error() {
					t.Fatalf("Expected error does not match\nexpected: %s\nreceived: %s", tc.expectErr, err.Error())
				}

				return
			}

			if tc.expectErr != "" {
				t.Fatalf("Expected to throw error\nexpected: %s\nreceived: <nil>", tc.expectErr)
			}

			if tc.options != nil {
				if *tc.options != *options {
					t.Fatalf("Options do not match\nexpected: %+v\nreceived: %+v", tc.options, options)
				}
			}
		})
	}
}

func TestOutput(t *testing.T) {
	tt := []struct {
		name                  string
		args                  []string
		expectedOutputContain string
	}{
		{"version_flag", []string{"app", "-version"}, "version"},
		{"help_flag", []string{"app", "-help"}, "Usage"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			if os.Getenv("EXIT_TEST") == "1" {
				oldArgs := os.Args
				os.Args = tc.args

				InstallFlags("version", os.Stdout)

				os.Args = oldArgs

				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=TestOutput/"+tc.name)
			cmd.Env = append(os.Environ(), "EXIT_TEST=1")
			output, _ := cmd.CombinedOutput()

			if e := strings.Contains(string(output[:]), tc.expectedOutputContain); !e {
				t.Fatalf("Output does not contain: %s\nreceived: %s", tc.expectedOutputContain, string(output[:]))
			}
		})
	}
}
