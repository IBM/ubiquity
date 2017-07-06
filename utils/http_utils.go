/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/gorilla/mux"
)

func ExtractErrorResponse(response *http.Response) error {
	errorResponse := resources.GenericResponse{}
	err := UnmarshalResponse(response, &errorResponse)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", errorResponse.Err)
}

func FormatURL(url string, entries ...string) string {
	base := url
	if !strings.HasSuffix(url, "/") {
		base = fmt.Sprintf("%s/", url)
	}
	suffix := ""
	for _, entry := range entries {
		suffix = path.Join(suffix, entry)
	}
	return fmt.Sprintf("%s%s", base, suffix)
}

func HttpExecuteUserAuth(httpClient *http.Client, logger *log.Logger, requestType string, requestURL string, user string, password string, rawPayload interface{}) (*http.Response, error) {
	payload, err := json.MarshalIndent(rawPayload, "", " ")
	if err != nil {
		logger.Printf("Internal error marshalling params %#v", err)
		return nil, fmt.Errorf("Internal error marshalling params")
	}

	if (user == "") {
		return nil, fmt.Errorf("Empty UserName passed")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Printf("Error in creating request %#v", err)
		return nil, fmt.Errorf("Error in creating request")
	}

        request.Header.Add("Content-Type","application/json")
        request.Header.Add("Accept","application/json")

        request.SetBasicAuth(user,password)
	return httpClient.Do(request)

}

func HttpExecute(httpClient *http.Client, logger *log.Logger, requestType string, requestURL string, rawPayload interface{}) (*http.Response, error) {
	payload, err := json.MarshalIndent(rawPayload, "", " ")
	if err != nil {
		logger.Printf("Internal error marshalling params %#v", err)
		return nil, fmt.Errorf("Internal error marshalling params")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Printf("Error in creating request %#v", err)
		return nil, fmt.Errorf("Error in creating request")
	}

	return httpClient.Do(request)
}

func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func Unmarshal(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalResponse(r *http.Response, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}
func UnmarshalDataFromRequest(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

type encoding int

const (
	encodePath encoding = 1 + iota

	encodePathSegment

	encodeHost

	encodeZone

	encodeUserPassword

	encodeQueryComponent

	encodeFragment
)

type EscapeError string

func (e EscapeError) Error() string {

	return "invalid URL escape " + strconv.Quote(string(e))

}

type InvalidHostError string

func (e InvalidHostError) Error() string {

	return "invalid character " + strconv.Quote(string(e)) + " in host name"

}

func PathUnescape(s string) (string, error) {

	return unescape(s, encodePathSegment)

}

func unescape(s string, mode encoding) (string, error) {

	// Count %, check that they're well-formed.

	n := 0

	hasPlus := false

	for i := 0; i < len(s); {

		switch s[i] {

		case '%':

			n++

			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {

				s = s[i:]

				if len(s) > 3 {

					s = s[:3]

				}

				return "", EscapeError(s)

			}

			// Per https://tools.ietf.org/html/rfc3986#page-21

			// in the host component %-encoding can only be used

			// for non-ASCII bytes.

			// But https://tools.ietf.org/html/rfc6874#section-2

			// introduces %25 being allowed to escape a percent sign

			// in IPv6 scoped-address literals. Yay.

			if mode == encodeHost && unhex(s[i+1]) < 8 && s[i:i+3] != "%25" {

				return "", EscapeError(s[i : i+3])

			}

			if mode == encodeZone {

				// RFC 6874 says basically "anything goes" for zone identifiers

				// and that even non-ASCII can be redundantly escaped,

				// but it seems prudent to restrict %-escaped bytes here to those

				// that are valid host name bytes in their unescaped form.

				// That is, you can use escaping in the zone identifier but not

				// to introduce bytes you couldn't just write directly.

				// But Windows puts spaces here! Yay.

				v := unhex(s[i+1])<<4 | unhex(s[i+2])

				if s[i:i+3] != "%25" && v != ' ' && shouldEscape(v, encodeHost) {

					return "", EscapeError(s[i : i+3])

				}

			}

			i += 3

		case '+':

			hasPlus = mode == encodeQueryComponent

			i++

		default:

			if (mode == encodeHost || mode == encodeZone) && s[i] < 0x80 && shouldEscape(s[i], mode) {

				return "", InvalidHostError(s[i : i+1])

			}

			i++

		}

	}

	if n == 0 && !hasPlus {

		return s, nil

	}

	t := make([]byte, len(s)-2*n)

	j := 0

	for i := 0; i < len(s); {

		switch s[i] {

		case '%':

			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])

			j++

			i += 3

		case '+':

			if mode == encodeQueryComponent {

				t[j] = ' '

			} else {

				t[j] = '+'

			}

			j++

			i++

		default:

			t[j] = s[i]

			j++

			i++

		}

	}

	return string(t), nil

}

func shouldEscape(c byte, mode encoding) bool {

	// §2.3 Unreserved characters (alphanum)

	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {

		return false

	}

	if mode == encodeHost || mode == encodeZone {

		// §3.2.2 Host allows

		//	sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="

		// as part of reg-name.

		// We add : because we include :port as part of host.

		// We add [ ] because we include [ipv6]:port as part of host.

		// We add < > because they're the only characters left that

		// we could possibly allow, and Parse will reject them if we

		// escape them (because hosts can't use %-encoding for

		// ASCII bytes).

		switch c {

		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':

			return false

		}

	}

	switch c {

	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)

		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)

		// Different sections of the URL allow a few of

		// the reserved characters to appear unescaped.

		switch mode {

		case encodePath: // §3.3

			// The RFC allows : @ & = + $ but saves / ; , for assigning

			// meaning to individual path segments. This package

			// only manipulates the path as a whole, so we allow those

			// last three as well. That leaves only ? to escape.

			return c == '?'

		case encodePathSegment: // §3.3

			// The RFC allows : @ & = + $ but saves / ; , for assigning

			// meaning to individual path segments.

			return c == '/' || c == ';' || c == ',' || c == '?'

		case encodeUserPassword: // §3.2.1

			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in

			// userinfo, so we must escape only '@', '/', and '?'.

			// The parsing of userinfo treats ':' as special so we must escape

			// that too.

			return c == '@' || c == '/' || c == '?' || c == ':'

		case encodeQueryComponent: // §3.4

			// The RFC reserves (so we must escape) everything.

			return true

		case encodeFragment: // §4.1

			// The RFC text is silent but the grammar allows

			// everything, so escape nothing.

			return false

		}

	}

	// Everything else must be escaped.

	return true

}

func unhex(c byte) byte {

	switch {

	case '0' <= c && c <= '9':

		return c - '0'

	case 'a' <= c && c <= 'f':

		return c - 'a' + 10

	case 'A' <= c && c <= 'F':

		return c - 'A' + 10

	}

	return 0

}

func ishex(c byte) bool {

	switch {

	case '0' <= c && c <= '9':

		return true

	case 'a' <= c && c <= 'f':

		return true

	case 'A' <= c && c <= 'F':

		return true

	}

	return false

}
