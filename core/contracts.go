package core

// Runner defines interface of a test runner.
type Runner interface {
	Run(requester Requester, testsuite *Testsuite) (noFailure bool, err error)
	Stop()
}

// Requester defines interface of testcase handler.
type Requester interface {
	Request(tc TestCase) (bool, error)
}

// TestResult defines interface to check if test has passed and
// to get string report.
type TestResult interface {
	Passed() bool
	String() string
}

// Testsuite hold all fields realted to testsuite and all testcases.
type Testsuite struct {
	Tests []TestCase `json:"tests"`
}

// TestCase specifies one test case.
type TestCase struct {
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Assertions Assertions        `json:"assertions"`
}

// Assertions describes expectations on each test case.
type Assertions struct {
	StatusCode int               `json:"statusCode"`
	Body       []string          `json:"body"`
	Headers    map[string]string `json:"headers"`
}
