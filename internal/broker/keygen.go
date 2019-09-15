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

package broker

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/emitter-io/emitter/internal/errors"
	"github.com/emitter-io/emitter/internal/security"
)

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(payload []byte) (response, bool) {
	// Deserialize the payload.
	message := keyGenRequest{}
	if err := json.Unmarshal(payload, &message); err != nil {
		return errors.ErrBadRequest, false
	}

	key, err := c.service.generateKey(message.Key, message.Channel, message.access(), message.expires())
	if err != nil {
		return err, false
	}

	// Success, return the response
	return &keyGenResponse{
		Status:  200,
		Key:     key,
		Channel: message.Channel,
	}, true
}

func (s *Service) generateKey(rawMasterKey string, channel string, access uint8, expires time.Time) (string, *errors.Error) {
	// Attempt to parse the key, this should be a master key
	masterKey, err := s.Cipher.DecryptKey([]byte(rawMasterKey))
	if err != nil || !masterKey.IsMaster() || masterKey.IsExpired() {
		return "", errors.ErrUnauthorized
	}

	// Attempt to fetch the contract using the key. Underneath, it's cached.
	contract, contractFound := s.contracts.Get(masterKey.Contract())
	if !contractFound {
		return "", errors.ErrNotFound
	}

	// Validate the contract
	if !contract.Validate(masterKey) {
		return "", errors.ErrUnauthorized
	}

	// Generate random salt
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt16))
	if err != nil {
		return "", errors.ErrServerError
	}

	// Create a key request
	key := security.Key(make([]byte, 24))
	key.SetSalt(uint16(n.Uint64()))
	key.SetMaster(masterKey.Master())
	key.SetContract(masterKey.Contract())
	key.SetSignature(masterKey.Signature())
	key.SetPermissions(access)
	key.SetExpires(expires)

	// Set the target and return an convert the error if it occurs
	if err := key.SetTarget(channel); err != nil {
		switch err {
		case security.ErrTargetInvalid:
			return "", errors.ErrTargetInvalid
		case security.ErrTargetTooLong:
			return "", errors.ErrTargetTooLong
		default:
			return "", errors.ErrServerError
		}
	}

	// Encrypt the final key
	out, err := s.Cipher.EncryptKey(key)
	if err != nil {
		return "", errors.ErrServerError
	}

	return out, nil
}

// ------------------------------------------------------------------------------------

type keygenForm struct {
	Key      string
	Channel  string
	TTL      int64
	Sub      bool
	Pub      bool
	Store    bool
	Load     bool
	Presence bool
	Extend   bool
	Response string
}

// perform type checking
func (f *keygenForm) parse(req *http.Request) bool {

	var ok bool
	var parsedOK = true

	canParseToBool := func(source string) (bool, bool) {
		value := req.FormValue(source)
		if value == "" {
			return false, true
		} else if value == "on" {
			return true, true
		} else if value == "off" {
			return false, true
		}

		return false, false
	}

	canParseToInt := func(source string) (int64, bool) {
		value := req.FormValue(source)
		if value == "" {
			return 0, true
		}

		b, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, false
		}
		return b, true
	}

	if f.Sub, ok = canParseToBool("sub"); !ok {
		parsedOK = false
	}

	if f.Pub, ok = canParseToBool("pub"); !ok {
		parsedOK = false
	}

	if f.Store, ok = canParseToBool("store"); !ok {
		parsedOK = false
	}

	if f.Load, ok = canParseToBool("load"); !ok {
		parsedOK = false
	}

	if f.Presence, ok = canParseToBool("presence"); !ok {
		parsedOK = false
	}

	if f.Extend, ok = canParseToBool("extend"); !ok {
		parsedOK = false
	}

	if f.TTL, ok = canParseToInt("ttl"); !ok {
		parsedOK = false
	}

	f.Key = req.FormValue("key")
	f.Channel = req.FormValue("channel")

	return parsedOK
}

// validate required fields
func (f *keygenForm) isValid() bool {
	if f.Key == "" {
		f.Response = "Missing SecretKey"
		return false
	}

	if f.Channel == "" {
		f.Response = "Missing Channel"
		return false
	}

	return true
}

func (f *keygenForm) expires() time.Time {
	if f.TTL == 0 {
		return time.Unix(0, 0)
	}

	return time.Now().Add(time.Duration(f.TTL) * time.Second).UTC()
}

func (f *keygenForm) access() uint8 {
	required := security.AllowNone
	if f.Sub {
		required |= security.AllowRead
	}
	if f.Pub {
		required |= security.AllowWrite
	}
	if f.Load {
		required |= security.AllowLoad
	}
	if f.Store {
		required |= security.AllowStore
	}
	if f.Presence {
		required |= security.AllowPresence
	}
	if f.Extend {
		required |= security.AllowExtend
	}

	return required
}

func handleKeyGen(s *Service) http.HandlerFunc {
	var fs http.FileSystem = assets
	f, err := fs.Open("keygen.html")
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	t, err := template.New("keygen").
		Funcs(template.FuncMap{
			"isChecked": func(checked bool) string {
				if checked {
					return "checked=\"true\""
				}
				return ""
			}}).
		Parse(string(b))
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := keygenForm{Sub: true}
		switch r.Method {
		case "GET":
		case "POST":
			ok := f.parse(r)
			if ok {
				if f.isValid() {
					key, err := s.generateKey(f.Key, f.Channel, f.access(), f.expires())
					if err != nil {
						f.Response = err.Error()
					} else {
						f.Response = fmt.Sprintf("channel: %s\nkey    : %s", f.Channel, key)
					}

				}
			} else {
				f.Response = "invalid arguments"
			}

		default:
			http.Error(w, http.ErrNotSupported.Error(), 405)
			return
		}

		err := t.Execute(w, f)
		if err != nil {
			log.Printf("template execute error: %s\n", err.Error())
			http.Error(w, "internal server error", 500)
		}
	}
}
