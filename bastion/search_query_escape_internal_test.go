package bastion

import (
	"fmt"
	"net/http"
	"testing"
)

// A name is the value of the q search parameter: unescaped, "a&b" was cut at the "&", "a+b" was
// read as "a b", and "#" cut the request, so the search before a creation or an import missed the
// object or found another one.
func TestSearchEscapesTheName(t *testing.T) {
	const name = "a&b+c#d e"
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		if r.URL.Path == "/api/"+VersionWallixAPI312+"/usergroups/" && r.URL.Query().Get("q") == `group_name=a\&b+c#d e` {
			fmt.Fprint(w, `[{"id":"group-id"}]`)

			return
		}
		fmt.Fprint(w, `[]`)
	})

	id, found, err := searchResourceUserGroup(t.Context(), name, client)
	if err != nil || !found || id != "group-id" {
		t.Errorf("search = (%q, %v, %v), want the group named %q", id, found, err, name)
	}
}

// The Bastion reads q as an expression: "&&" and "||" are operators and "\" escapes the next
// character. Checked against a Bastion 12.4.1, each escaped value below finds exactly the object
// with that name.
func TestSearchQueryEscapesTheExpression(t *testing.T) {
	tests := map[string]string{
		"my_user_group": `q=group_name=my_user_group`,
		"a&&b":          `q=group_name=a%5C%26%5C%26b`,
		"a||b":          `q=group_name=a%5C%7C%5C%7Cb`,
		`back\slash`:    `q=group_name=back%5C%5Cslash`,
		"(a)=b,c":       `q=group_name=%5C%28a%5C%29%5C%3Db%5C%2Cc`,
		`a"b'c`:         `q=group_name=a%5C%22b%5C%27c`,
		"a!b<c>*":       `q=group_name=a%5C%21b%5C%3Cc%5C%3E%5C%2A`,
		"a b+c#d":       `q=group_name=a+b%2Bc%23d`,
	}
	for name, want := range tests {
		if got := searchQuery("group_name", name); got != want {
			t.Errorf("searchQuery(%q) = %s, want %s", name, got, want)
		}
	}
}
