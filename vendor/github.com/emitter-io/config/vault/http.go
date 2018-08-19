// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package vault

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

// httpHeader represents a header with a value attached.
type httpHeader struct {
	Header string
	Value  string
}

// newHTTPHeader builds an HTTP header with a value.
func newHTTPHeader(header, value string) httpHeader {
	return httpHeader{Header: header, Value: value}
}

// httpGet is a utility function which issues an HTTP httpGet on a specified URL. The encoding is JSON.
var httpGet = func(url string, output interface{}, headers ...httpHeader) error {
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

	return unmarshalJSON(resp.Body, output)
}

// httpPost is a utility function which marshals and issues an HTTP post on a specified URL. The
// encoding is JSON.
var httpPost = func(url string, body interface{}, output interface{}, headers ...httpHeader) error {
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

	return unmarshalJSON(resp.Body, output)
}

// unmarshalJSON unmarshals the given io.Reader pointing to a JSON, into a desired object
var unmarshalJSON = func(r io.Reader, out interface{}) error {
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
