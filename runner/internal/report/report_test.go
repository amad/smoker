package report_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amad/smoker/runner/internal/report"
)

func TestReport(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		report         *report.TestReport
		expectedStatus bool
		expectedString string
	}{
		{"passed", &report.TestReport{1, "a", true, nil, time.Duration(1) * time.Second}, true, "PASS: testcase #1 \"a\" (1.00s)"},
		{"failed", &report.TestReport{2, "b", false, errors.New("reason"), time.Duration(2) * time.Second}, false, "FAIL: testcase #2 \"b\" reason (2.00s)"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.report.Passed() != tc.expectedStatus {
				t.Fatalf("Status does not match\nexpected: %t\nreceived: %t", tc.expectedStatus, tc.report.Passed())
			}

			if tc.expectedString != tc.report.String() {
				t.Fatalf("Report string does not match\nexpected: %s\nreceived: %s", tc.expectedString, tc.report.String())
			}
		})
	}

}
