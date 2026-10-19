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
func deleteDone(code int) bool {
	return code == http.StatusOK || code == http.StatusNoContent || code == http.StatusNotFound
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
