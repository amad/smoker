# Smoker

![Tests](https://github.com/amad/smoker/workflows/Tests/badge.svg?branch=master)
[![Build Status](https://travis-ci.org/amad/smoker.svg?branch=master)](https://travis-ci.org/amad/smoker)
[![Go Report Card](https://goreportcard.com/badge/github.com/amad/smoker)](https://goreportcard.com/report/github.com/amad/smoker)
[![codecov](https://codecov.io/gh/amad/smoker/branch/master/graph/badge.svg)](https://codecov.io/gh/amad/smoker)
[![License: MPL2.0](https://img.shields.io/badge/license-MPL2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Fast smoke-testing tool for APIs and WEB with concurrency support.

Describe test cases in JSON format. Each test case can define URL, HTTP method, request headers, and request body. Smoker makes assertions base on the response status-code and response body.

## Installation

### From source

```bash
go get github.com/amad/smoker/cmd/smoker
```

### Docker

```bash
docker pull stunt/smoker:latest
```

### Download pre-compiled

```bash
# Find latest version: https://github.com/amad/smoker/releases
curl -o smoker.tar.gz -L https://github.com/amad/smoker/releases/download/v0.2.0/smoker_0.2.0_linux_amd64.tar.gz
tar -zxvf smoker.tar.gz --strip-components 1
chmod +x ./smoker
./smoker -version
```

### Requirements

- Go 1.13

## Usage

Run a testsuite:

```txt
smoker -testsuite smoke-api.json
```

Run with 15 workers and set global timeout to 5 seconds:

```txt
smoker -testsuite smoke-api.json -workers 15 -timeout 5
```

Run with `-stop-on-failure` flag to stop execution if any test-case fails:

```txt
smoker -testsuite smoke-api.json -stop-on-failure
```

```txt
Usage: smoker [options...]

Example:
  smoker -testsuite api.json
  smoker -testsuite web.json -workers 10 -timeout 5 -stop-on-failure

Options:
  -testsuite        Testsuite file in JSON format to read test cases.
  -workers          Number of workers to send requests concurrently. (accepts integer value >= 1. Default is 1. 0 is not allowed)
  -timeout          Set timeout per request in seconds. (accepts integer value >= 1. Default is 10. 0 is not allowed)
  -stop-on-failure  Stop execution upon first error or failure.
  -version          Prints the version and exits.
```

## How to describe test cases

A testsuite is a JSON file with the following structure. It accept an array of test cases, and you can have hundreds of test cases on each testsuite file. There is no limit on number of test cases per testsuite file.

```txt
{
  "tests": [
    // test cases here
  ]
}
```

A test case can have all the following parameters to give you more control on what you want to test. But, most of these fields are optional. You can find more examples below.

```json
{
  "tests": [
    {
      "name": "A test case with all parameters",
      "url": "https://api.github.com",
      "method": "post",
      "body": "{\"test\":1}",
      "headers": {
        "content-type": "application/json"
      },
      "assertions": {
        "statusCode": 200,
        "body": [
          "github",
          "[a-z]"
        ],
        "header": {
          "Content-Type": "application/json"
        }
      }
    }
  ]
}
```

Minimum requirement for a test case is to have a `name` field, and a `url` field. All the other fields are optional.
The default request method is `GET` and the default assertion is to match `200` HTTP status code.

```json
{
  "tests": [
    {
      "name": "Assert github.com/amad/smoker returns 200",
      "url": "https://github.com/amad/smoker"
    },
    {
      "name": "Assert Github is OK",
      "url": "https://github.com"
    }
  ]
}
```

You can optionally set the HTTP method, request body, and request headers as well. All these fields are optional.

Example:

```json
{
  "tests": [
    {
      "name": "Send a POST request with header and body",
      "url": "https://api.github.com",
      "method": "post",
      "body": "{\"test\":1}",
      "headers": {
        "content-type": "application/json"
      }
    }
  ]
}
```

You can make assertion on HTTP status code, and also assert whether the response contains any match of the provided regular expression or simple string. You can also make assertion on response header.

Test case fails if the HTTP status code does not match, Or any of the assertion in body do not match.

The `assertions.statusCode` field accepts a HTTP status code to check. The default value for this field is `200`.

The `assertions.body` field accepts an array of strings that can contain simple string or regex. When this field isn't provided, Smoker does not make any assertions on response body.

The `assertions.header` field accepts a map of strings. When this field isn't provided, Smoker does not make any assertions on response header. You can use regular expression to match header value. But, you must provide full header name.

Example:

```json
{
  "tests": [
    {
      "name": "Multiple assertions on response body and one assertion on status code",
      "url": "https://github.com/amad/NotFoundRepo",
      "assertions": {
        "statusCode": 404,
        "body": [
          "Github",
          "Page not found",
          "[a-z]"
        ],
        "header": {
          "Content-Type": "application/json",
          "X-Requestid": "*"
        }
      }
    }
  ]
}
```
