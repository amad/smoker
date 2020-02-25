package loader_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/amad/smoker/core"
	"github.com/amad/smoker/loader"
)

// LoadTestsuite loads testsuite from a JSON file
func TestLoadTestsuite(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name      string
		filename  string
		expectRes *core.Testsuite
		expectErr string
	}{
		{"load a testsuite", "./testdata/suite1.json", &core.Testsuite{Tests: []core.TestCase{{Name: "test case 1", URL: "https://github.com/amad/smoker"}}}, ""},
		{"should error on invalid file type", "./testdata/textfile", &core.Testsuite{}, "unable to parse config file"},
		{"should error on wrong path", "./testdata/notfound.json", &core.Testsuite{}, "unable to open config file"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := loader.LoadTestsuite(tc.filename)

			if err != nil {
				if tc.expectErr == "" {
					t.Fatalf("Unexpected error\nexpected: <nil>\nreceived: %s", err.Error())
				}

				if ok := strings.Contains(err.Error(), tc.expectErr); !ok {
					t.Fatalf("Expected error does not match\nexpected: %s\nreceived: %s", tc.expectErr, err.Error())
				}

				return
			}

			if tc.expectErr != "" {
				t.Fatalf("Expected to throw error\nexpected: %s\nreceived: <nil>", tc.expectErr)
			}

			if !reflect.DeepEqual(tc.expectRes, res) {
				t.Fatalf("Response does not match\nexpected: %+v\nreceived: %+v", *tc.expectRes, *res)
			}
		})
	}
}
