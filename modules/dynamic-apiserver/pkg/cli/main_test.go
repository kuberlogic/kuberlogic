/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func makeTestClient(code int, response interface{}) *http.Client {
	data, _ := json.Marshal(response)
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: code,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(data)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	return client
}
