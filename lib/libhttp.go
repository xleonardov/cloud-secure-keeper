// Package lib provides http related library functions.
package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// BasicRealm is used when setting the WWW-Authenticate response header.
var BasicRealm = "Authorization Required"

// ParseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func ParseBasicAuth(auth string) (username string, password string, ok bool) {
	if !strings.HasPrefix(auth, "Basic ") {
		return
	}
	c, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// BasicAuthUnauthorized denies authentication.
func BasicAuthUnauthorized(w http.ResponseWriter, err error) {
	message := "Not Authorized."

	if err != nil {
		message += fmt.Sprintf(" Error: %v", err)
	}

	w.Header().Set("WWW-Authenticate", "Basic realm=\""+BasicRealm+"\"")
	http.Error(w, message, http.StatusUnauthorized)
}