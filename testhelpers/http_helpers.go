// Package testhelpers provides utility functions for testing.
package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewJSONRequest creates a new HTTP request with JSON body.
func NewJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// ExecuteRequest executes an HTTP request and returns the response recorder.
func ExecuteRequest(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// ReadResponseBody reads and unmarshals the response body into the provided struct.
func ReadResponseBody(rr *httptest.ResponseRecorder, dest interface{}) error {
	return json.NewDecoder(rr.Body).Decode(dest)
}
