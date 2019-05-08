package broker

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	conf "github.com/emitter-io/config"
	"github.com/emitter-io/emitter/internal/config"
	"github.com/stretchr/testify/assert"
)

var keygenTestLicense = "zT83oDV0DWY5_JysbSTPTDr8KB0AAAAAAAAAAAAAAAI"
var keygenTestSecret = "kBCZch5re3Ue-kpG1Aa8Vo7BYvXZ3UwR"

func setup(t *testing.T) (http.HandlerFunc, func()) {
	cfg := config.NewDefault().(*config.Config)
	cfg.License = keygenTestLicense
	cfg.TLS = &conf.TLSConfig{}

	broker, svcErr := NewService(context.Background(), cfg)
	assert.NoError(t, svcErr)

	teardown := func() {
		broker.Close()
	}

	return handleKeyGen(broker), teardown
}

var keyGenResponseM = regexp.MustCompile(`(?s)<pre id="keygenResponse">(?P<response>.*)</pre>`)

func TestRenderKeyGenPage(t *testing.T) {

	handler, teardown := setup(t)
	defer teardown()

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

	handler, teardown := setup(t)
	defer teardown()

	req := httptest.NewRequest("PUT", "https://emitter.io/keygen", nil)
	w := httptest.NewRecorder()

	// act
	handler(w, req)

	// assert
	assert.Equal(t, 405, w.Code)
}

func TestGenerateKey(t *testing.T) {

	handler, teardown := setup(t)
	defer teardown()

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
		testCase{
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
		testCase{
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
		testCase{
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
		testCase{
			Scenario:                 "Request with invalid TTL",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			TTL:                      "ERR",
			ExpectedResponseContains: "invalid arguments",
		},
		testCase{
			Scenario:                 "Request with invalid permission argument",
			Key:                      keygenTestSecret,
			Channel:                  "bar/",
			PermissionSub:            "ERR",
			ExpectedResponseContains: "invalid arguments",
		},
		testCase{
			Scenario:                 "Pass missing secret",
			Key:                      "",
			Channel:                  "bar/",
			PermissionSub:            "off",
			ExpectedResponseContains: "Missing SecretKey",
		},
		testCase{
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
