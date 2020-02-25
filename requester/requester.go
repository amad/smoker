package requester

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/amad/smoker/core"
	"github.com/google/uuid"
)

// NewRequester creates and returns new a Requester.
func NewRequester(timeout time.Duration, userAgent string) *Requester {
	client := &http.Client{
		Timeout: timeout,
	}

	return &Requester{
		client:    client,
		userAgent: userAgent,
	}
}

// Requester handles HTTP requests.
type Requester struct {
	client    *http.Client
	userAgent string
}

// Request method uses HTTP package to send request and verifies if the
// response matches test case expectations.
func (r *Requester) Request(tc core.TestCase) (bool, error) {
	if tc.Name == "" {
		return false, errors.New("does not have name field")
	}

	if tc.URL == "" {
		return false, errors.New("does not have url field")
	}

	if tc.Method != "" {
		tc.Method = strings.ToUpper(tc.Method)
	} else {
		tc.Method = http.MethodGet
	}

	if tc.Assertions.StatusCode == 0 {
		tc.Assertions.StatusCode = http.StatusOK
	}

	req, err := http.NewRequest(tc.Method, tc.URL, nil)
	if err != nil {
		return false, fmt.Errorf("could not create request: %w", err)
	}

	id := uuid.New().String()
	req.Header.Set("Request-Id", id)
	req.Header.Set("User-Agent", r.userAgent)

	for name, value := range tc.Headers {
		req.Header.Set(name, value)
	}

	if tc.Body != "" {
		req.Body = ioutil.NopCloser(bytes.NewReader([]byte(tc.Body)))
	}

	res, err := r.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != tc.Assertions.StatusCode {
		return false, fmt.Errorf("expected status-code: %d received: %d", tc.Assertions.StatusCode, res.StatusCode)
	}

	if len(tc.Assertions.MatchInBody) != 0 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return false, fmt.Errorf("unable to read the response body with with error: %w", err)
		}

		bodyStr := string(body)

		for _, matchInBody := range tc.Assertions.MatchInBody {
			res, err := regexp.MatchString(matchInBody, bodyStr)
			if err != nil || !res {
				return false, fmt.Errorf("can not match /%s/ in response body", matchInBody)
			}
		}
	}

	return true, nil
}
