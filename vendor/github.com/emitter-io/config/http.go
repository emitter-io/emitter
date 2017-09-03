package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
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

// newHttpHeader builds an HTTP header with a value.
func newHttpHeader(header, value string) httpHeader {
	return httpHeader{Header: header, Value: value}
}

// httpFile downloads a file from HTTP
var httpFile = func(url string) (*os.File, error) {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]

	output, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if _, err := io.Copy(output, response.Body); err != nil {
		return nil, err
	}

	return output, nil
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
