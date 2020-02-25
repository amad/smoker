package report

import (
	"fmt"
	"time"
)

// The TestReport holds results of a test case.
type TestReport struct {
	Index    int
	Name     string
	Status   bool
	Err      error
	Duration time.Duration
}

// String method returns the test result as string.
func (r *TestReport) String() string {
	if !r.Passed() {
		return fmt.Sprintf("FAIL: testcase #%d \"%s\" %s (%.2fs)", r.Index, r.Name, r.Err, r.Duration.Seconds())
	}

	return fmt.Sprintf("PASS: testcase #%d \"%s\" (%.2fs)", r.Index, r.Name, r.Duration.Seconds())
}

// Passed method checks if test result was successful.
func (r *TestReport) Passed() bool {
	return r.Status
}
