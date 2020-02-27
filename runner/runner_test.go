package runner

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/amad/smoker/core"
)

func TestNewRunner(t *testing.T) {
	t.Parallel()

	var workers = 5
	var buffer *bytes.Buffer

	NewRunner(workers, time.Second, false, buffer, buffer)
}

func newTestRunner(workers int, timeout int, stopOnFailure bool) *Runner {
	var buffer bytes.Buffer

	return NewRunner(workers, time.Duration(timeout)*time.Second, stopOnFailure, &buffer, &buffer)
}

func TestPrintfOutAndPrintfErrOut(t *testing.T) {
	expectedOutput := "test msg\n"
	expectedErrOutput := "test error\n"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	r := NewRunner(1, time.Second, false, &stdout, &stderr)

	r.printfErrOut("test %s", "error")
	r.printfOut("test %s", "msg")

	if stdout.String() != expectedOutput {
		t.Fatalf("stdout output does not match\nexpected: %s\nreceived: %s", expectedOutput, stdout.String())
	}

	if stderr.String() != expectedErrOutput {
		t.Fatalf("stderr output does not match\nexpected: %s\nreceived: %s", expectedErrOutput, stderr.String())
	}
}

func TestGetPoolsize(t *testing.T) {
	tt := []struct {
		name           string
		numWorkers     int
		numTestcases   int
		expectPoolSize int
	}{
		{"one worker", 1, 5, 1},
		{"pool size is determined by num of workers", 2, 5, 2},
		{"pool size can not be higher than test cases", 5, 3, 3},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			r := NewRunner(tc.numWorkers, time.Second, false, &buffer, &buffer)
			ts := &core.Testsuite{}
			for n := 0; n < tc.numTestcases; n++ {
				ts.Tests = append(ts.Tests, core.TestCase{})
			}

			poolSize := r.getPoolsize(ts)

			if poolSize != tc.expectPoolSize {
				t.Fatalf("getPoolsize error\n expected: %d\nreceived: %d", tc.expectPoolSize, poolSize)
			}
		})
	}
}

type testRequester struct{}

func (r *testRequester) Request(tc core.TestCase) (bool, error) {
	if tc.Name == "fail" {
		return false, errors.New("testRequester fake failure")
	}

	return true, nil
}

func TestRunnerWithEmptySuite(t *testing.T) {
	requester := &testRequester{}
	runner := newTestRunner(1, 1, false)

	_, err := runner.Run(requester, &core.Testsuite{})
	if err == nil {
		t.Fatal("Expected to throw error when testsuite is empty")
	}
}

func TestRunner(t *testing.T) {
	tt := []struct {
		name              string
		testsuite         *core.Testsuite
		stopOnFailure     bool
		expectPassedCount int
		expectFailedCount int
	}{
		{
			"all cases pass",
			&core.Testsuite{Tests: []core.TestCase{{}, {}, {}}},
			false,
			3,
			0,
		},
		{
			"one case should fail",
			&core.Testsuite{Tests: []core.TestCase{{Name: "fail"}, {}, {}}},
			false,
			2,
			1,
		},
		{
			"all cases fail",
			&core.Testsuite{Tests: []core.TestCase{{Name: "fail"}, {Name: "fail"}, {Name: "fail"}}},
			false,
			0,
			3,
		},
		{
			"stop on failure",
			&core.Testsuite{Tests: []core.TestCase{{Name: "fail"}, {}, {}}},
			true,
			0,
			1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			requester := &testRequester{}
			runner := newTestRunner(1, 1, tc.stopOnFailure)

			ok, _ := runner.Run(requester, tc.testsuite)

			if tc.expectFailedCount > 0 && ok {
				t.Fatalf("Expected report to have failed test cases")
			}

			if buffer := runner.stdout.(*bytes.Buffer); buffer.String() == "" {
				t.Fatalf("Runner is not writing output")
			}

			expectReport(t, &runner.reports, tc.expectPassedCount, tc.expectFailedCount)
		})
	}
}

func expectReport(t *testing.T, reports *[]core.TestResult, expectPassedCount int, expectFailedCount int) {
	var cp = 0
	var cf = 0
	for _, report := range *reports {
		if report.Passed() {
			cp = cp + 1
		} else {
			cf = cf + 1
		}
	}

	if cp != expectPassedCount {
		t.Fatalf("Expected report to have %d passed tests, got %d", expectPassedCount, cp)
	}

	if cf != expectFailedCount {
		t.Fatalf("Expected report to have %d failed tests, got %d", expectFailedCount, cf)
	}
}
