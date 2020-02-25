package loader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/amad/smoker/core"
)

// LoadTestsuite loads testsuite from a JSON file.
func LoadTestsuite(filename string) (*core.Testsuite, error) {
	var testsuite core.Testsuite

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &testsuite, fmt.Errorf("unable to open config file: %w", err)
	}

	err = json.Unmarshal(contents, &testsuite)
	if err != nil {
		return &testsuite, fmt.Errorf("unable to parse config file: %w", err)
	}

	return &testsuite, nil
}
