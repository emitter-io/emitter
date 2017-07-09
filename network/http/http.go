package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// DefaultClient used for http with a shorter timeout.
var defaultClient = &http.Client{
	Timeout: 5 * time.Second,
}

// HeaderValue represents a header with a value attached.
type HeaderValue struct {
	Header string
	Value  string
}

// NewHeader builds an HTTP header with a value.
func NewHeader(header, value string) HeaderValue {
	return HeaderValue{Header: header, Value: value}
}

// Get is a utility function which issues an HTTP Get on a specified URL. The encoding is JSON.
func Get(url string, output interface{}, headers ...HeaderValue) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Set the headers provided
	for _, h := range headers {
		req.Header.Set(h.Header, h.Value)
	}

	// Issue the request
	resp, err := defaultClient.Do(req)
	if err != nil {
		return err
	}

	return UnmarshalJSON(resp.Body, output)
}

// Post is a utility function which marshals and issues an HTTP post on a specified URL. The
// encoding is JSON.
func Post(url string, body interface{}, output interface{}, headers ...HeaderValue) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Build a new request
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	// Set the headers provided
	for _, h := range headers {
		req.Header.Set(h.Header, h.Value)
	}

	// Issue the request
	resp, err := defaultClient.Do(req)
	if err != nil {
		return err
	}

	return UnmarshalJSON(resp.Body, output)
}

// UnmarshalJSON unmarshals the given io.Reader pointing to a JSON, into a desired object
func UnmarshalJSON(r io.Reader, out interface{}) error {
	if r == nil {
		return errors.New("'io.Reader' being decoded is nil")
	}

	if out == nil {
		return errors.New("output parameter 'out' is nil")
	}

	// Decode the json
	dec := json.NewDecoder(r)
	return dec.Decode(out)
}
