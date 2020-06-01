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
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/emitter-io/emitter/internal/security"
)

// HTTP creates a new HTTP handler which can be used to serve HTTP keygen page.
func (s *Service) HTTP() http.HandlerFunc {
	t, err := template.New("keygen").
		Funcs(template.FuncMap{
			"isChecked": func(checked bool) string {
				if checked {
					return "checked=\"true\""
				}
				return ""
			}}).
		Parse(string(indexPage))
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
					key, err := s.CreateKey(f.Key, f.Channel, f.access(), f.expires())
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

		if err := t.Execute(w, f); err != nil {
			log.Printf("template execute error: %s\n", err.Error())
			http.Error(w, "internal server error", 500)
		}
	}
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
