/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

package keygen

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/emitter-io/emitter/internal/provider/contract"
	"github.com/emitter-io/emitter/internal/provider/usage"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/stretchr/testify/assert"
)

const keygenTestLicense = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI:1"
const keygenTestSecret = "kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR"

func newTestProvider(t *testing.T) *Service {
	l, err := license.Parse(keygenTestLicense)
	assert.NoError(t, err)

	cipher, err := l.Cipher()
	assert.NoError(t, err)

	provider := contract.NewSingleContractProvider(l, usage.NewNoop())
	return New(cipher, provider, &authorizer{cipher, provider})
}

var keyGenResponseM = regexp.MustCompile(`(?s)<pre id="keygenResponse">(?P<response>.*)</pre>`)

func TestRenderKeyGenPage(t *testing.T) {
	p := newTestProvider(t)
	handler := p.HTTP()

	req := httptest.NewRequest("GET", "https://emitter.io/keygen", nil)
	w := httptest.NewRecorder()

	// act
	handler(w, req)
	content, err := ioutil.ReadAll(w.Body)
	if err != nil {
		log.Fatal(err)
	}

	// assert
	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, content)
}

func TestHeadRequest(t *testing.T) {
	p := newTestProvider(t)
	handler := p.HTTP()

	req := httptest.NewRequest("PUT", "https://emitter.io/keygen", nil)
	w := httptest.NewRecorder()

	// act
	handler(w, req)

	// assert
	assert.Equal(t, 405, w.Code)
}

func Test_HTTP(t *testing.T) {
	p := newTestProvider(t)
	handler := p.HTTP()

	type testCase struct {
		Scenario                 string
		Key                      string
		Channel                  string
		TTL                      string
		PermissionSub            string
		PermissionPub            string
		PermissionStore          string
		PermissionLoad           string
		PermissionPresence       string
		PermissionExtend         string
		Response                 string
		ExpectedResponseContains string
	}

	testCases := []testCase{
		{
			Scenario:                 "Request with valid arguments",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			TTL:                      "300",
			PermissionSub:            "on",
			PermissionPub:            "on",
			PermissionLoad:           "on",
			PermissionStore:          "on",
			PermissionPresence:       "on",
			PermissionExtend:         "on",
			ExpectedResponseContains: "key    :",
		},
		{
			Scenario:                 "Request with empty valid arguments",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			TTL:                      "",
			PermissionSub:            "",
			PermissionPub:            "",
			PermissionLoad:           "",
			PermissionStore:          "",
			PermissionPresence:       "",
			PermissionExtend:         "",
			ExpectedResponseContains: "key    :",
		},
		{
			Scenario:                 "Request with invalid arguments",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			TTL:                      "bad",
			PermissionSub:            "bad",
			PermissionPub:            "bad",
			PermissionLoad:           "bad",
			PermissionStore:          "bad",
			PermissionPresence:       "bad",
			PermissionExtend:         "bad",
			ExpectedResponseContains: "invalid arguments",
		},
		{
			Scenario:                 "Request with invalid TTL",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			TTL:                      "ERR",
			ExpectedResponseContains: "invalid arguments",
		},
		{
			Scenario:                 "Request with invalid permission argument",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			PermissionSub:            "ERR",
			ExpectedResponseContains: "invalid arguments",
		},
		{
			Scenario:                 "Pass missing secret",
			Key:                      "",
			Channel:                  "bar/",
			PermissionSub:            "off",
			ExpectedResponseContains: "Missing SecretKey",
		},
		{
			Scenario:                 "Pass missing secret",
			Key:                      keygenTestSecret,
			Channel:                  "",
			PermissionSub:            "on",
			ExpectedResponseContains: "Missing Channel",
		},
	}

	for _, c := range testCases {

		data := url.Values{}
		data.Set("key", c.Key)
		data.Set("channel", c.Channel)
		data.Set("ttl", c.TTL)
		data.Set("sub", c.PermissionSub)
		data.Set("pub", c.PermissionPub)
		data.Set("store", c.PermissionStore)
		data.Set("load", c.PermissionLoad)
		data.Set("presence", c.PermissionPresence)
		data.Set("extend", c.PermissionExtend)

		req, _ := http.NewRequest("POST", "https://emitter.io/keygen", strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		w := httptest.NewRecorder()

		// act
		handler(w, req)
		content, err := ioutil.ReadAll(w.Body)
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, content)

		response := strings.TrimSpace(keyGenResponseM.FindStringSubmatch(string(content))[1])
		assert.Contains(t, response, c.ExpectedResponseContains, c.Scenario)
	}
}
