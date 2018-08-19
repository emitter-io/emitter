// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package config

import (
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
