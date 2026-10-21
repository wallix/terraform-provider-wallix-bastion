package bastion

import (
	"net/http"
	"net/url"
	"strings"
)

// searchQuery returns the q parameter searching for the objects whose field is exactly value.
// The Bastion reads q as an expression where "&&" and "||" are operators and "\" followed by any
// character is that character: unescaped, a name with "&&" or "||" failed with 400 "Invalid
// expression" and a "\" was dropped, so the object was not found.
func searchQuery(field, value string) string {
	escaped := strings.NewReplacer(
		`\`, `\\`, `&`, `\&`, `|`, `\|`, `(`, `\(`, `)`, `\)`, `=`, `\=`, `!`, `\!`,
		`<`, `\<`, `>`, `\>`, `,`, `\,`, `*`, `\*`, `"`, `\"`, `'`, `\'`,
	).Replace(value)

	return "q=" + field + "=" + url.QueryEscape(escaped)
}

// deleteDone reports whether the response to a DELETE means that the object is gone: deleted now
// (200, 204) or already deleted outside of Terraform (404). Terraform then drops it from state;
// returning the 404 as an error failed the destroy, and left the objects depending on it.
// The Bastion also answers 404 to an unknown API version or route: the object may still exist, so
// that 404 stays an error instead of dropping the object from state.
func deleteDone(code int, body string) bool {
	switch code {
	case http.StatusOK, http.StatusNoContent:
		return true
	case http.StatusNotFound:
		return !strings.Contains(body, "API version") && !strings.Contains(body, "This resource does not exist")
	default:
		return false
	}
}

type jsonRestriction struct {
	Action      string `json:"action"`
	Rules       string `json:"rules"`
	SubProtocol string `json:"subprotocol"`
}

type jsonCredential struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}
