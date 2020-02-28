package requester

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/amad/smoker/core"
)

var expectedUserAgent = "test-user-agent"
var expectedTimeout = time.Second

func TestNewRequester(t *testing.T) {
	t.Parallel()

	r := NewRequester(expectedTimeout, expectedUserAgent)

	if r.userAgent != expectedUserAgent {
		t.Fatalf("Expected to set correct user agent %s but received %s", expectedUserAgent, r.userAgent)
	}

	if r.client.Timeout != expectedTimeout {
		t.Fatalf("Expected to set timeout %d but received %d", expectedTimeout, r.client.Timeout)
	}
}

func TestRequest(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		tc             core.TestCase
		mockStatusCode int
		mockResBody    string
		mockResHeader  map[string]string
		expectErr      string
	}{
		{
			name: "OK",
			tc: core.TestCase{
				Name:    "test",
				URL:     "example.com",
				Method:  "get",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    "OK",
				Assertions: core.Assertions{
					StatusCode: 200,
					Body:       []string{"OK"},
				},
			},
			mockStatusCode: 200,
			mockResBody:    "OK",
			expectErr:      "",
		},
		{
			name: "name is required",
			tc: core.TestCase{
				Name: "",
			},
			mockStatusCode: 200,
			mockResBody:    "OK",
			expectErr:      "does not have name field",
		},
		{
			name: "url is required",
			tc: core.TestCase{
				Name: "test",
			},
			mockStatusCode: 200,
			mockResBody:    "OK",
			expectErr:      "does not have url field",
		},
		{
			name: "default method is get",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
			},
			mockStatusCode: 200,
		},
		{
			name: "errors when does not match status code",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
			},
			mockStatusCode: 500,
			expectErr:      "expected status-code: 200 received: 500",
		},
		{
			name: "errors when does not match body",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
				Assertions: core.Assertions{
					Body: []string{"OK"},
				},
			},
			mockStatusCode: 200,
			mockResBody:    "something else",
			expectErr:      "can not match /OK/ in response body",
		},
		{
			name: "can match body with regex",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
				Assertions: core.Assertions{
					Body: []string{"^[0-9]{3}-[0-9]{5}$"},
				},
			},
			mockStatusCode: 200,
			mockResBody:    "123-45678",
		},
		{
			name: "errors when can not match header",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
				Assertions: core.Assertions{
					Header: map[string]string{"Content-Type": "application/json"},
				},
			},
			mockStatusCode: 200,
			mockResHeader:  map[string]string{"Content-Type": "text/html"},
			expectErr:      "expected response header Content-Type:application/json received Content-Type:text/html",
		},
		{
			name: "errors when expected header not found",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
				Assertions: core.Assertions{
					Header: map[string]string{"Content-Type": "application/json"},
				},
			},
			mockStatusCode: 200,
			expectErr:      "unable to find response header Content-Type",
		},
		{
			name: "can match header",
			tc: core.TestCase{
				Name: "test",
				URL:  "example.com",
				Assertions: core.Assertions{
					Header: map[string]string{"access-control-allow-origin": "*", "content-length": "[0-9]+"},
				},
			},
			mockStatusCode: 200,
			mockResHeader:  map[string]string{"access-control-allow-origin": "*", "content-length": "23432"},
		},
	}

	for _, item := range tt {
		t.Run(item.tc.Name, func(t *testing.T) {
			mockClient := newTestClient(func(req *http.Request) *http.Response {
				if item.tc.Method != "" && req.Method != strings.ToUpper(item.tc.Method) {
					t.Fatalf("Request method does not match\nexpected: %s\nreceived: %s", strings.ToUpper(item.tc.Method), req.Method)
				} else if item.tc.Method == "" && req.Method != "GET" {
					t.Fatalf("Request method does not match\nexpected: %s\nreceived: %s", "GET", req.Method)
				}

				if req.URL.String() != item.tc.URL {
					t.Fatalf("Request URL does not match\nexpected: %s\nreceived: %s", item.tc.URL, req.URL.String())
				}

				if item.tc.Body != "" {
					if b, _ := ioutil.ReadAll(req.Body); string(b) != item.tc.Body {
						t.Fatalf("Request body does not match\nexpected: %s\nreceived: %s", item.tc.Body, string(b))
					}
				}

				if userAgent, ok := req.Header["User-Agent"]; !ok {
					t.Fatal("Expected to send User-Agent header but got <nil>")

					if userAgent[0] != expectedUserAgent {
						t.Fatalf("Header User-Agent does not match\nexpected: %s\nreceived: %s", expectedUserAgent, userAgent[0])
					}
				}

				if _, ok := req.Header["Request-Id"]; !ok {
					t.Fatal("Expected to send Request-Id header but got <nil>")
				}

				for headerName, headerValue := range item.tc.Headers {
					if val, ok := req.Header[headerName]; !ok || headerValue != val[0] {
						t.Fatalf("Header %s does not match\nexpected: %s\nreceived: %s", headerName, headerValue, val[0])
					}
				}

				resHeader := make(http.Header)

				for hn, hv := range item.mockResHeader {
					resHeader.Add(hn, hv)
				}

				return &http.Response{
					StatusCode: item.mockStatusCode,
					Body:       ioutil.NopCloser(bytes.NewBufferString(item.mockResBody)),
					Header:     resHeader,
				}
			})
			requester := &Requester{
				mockClient,
				expectedUserAgent,
			}

			_, err := requester.Request(item.tc)

			if err != nil {
				if item.expectErr == "" {
					t.Fatalf("Unexpected error\nexpected: <nil>\nreceived: %s", err.Error())
				}

				if ok := strings.Contains(err.Error(), item.expectErr); !ok {
					t.Fatalf("Expected error does not match\nexpected: %s\nreceived: %s", item.expectErr, err.Error())
				}

				return
			}

			if item.expectErr != "" {
				t.Fatalf("Expected to throw error\nexpected: %s\nreceived: <nil>", item.expectErr)
			}
		})
	}
}

type roundTripFunc func(r *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}
