package bastion

import (
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
