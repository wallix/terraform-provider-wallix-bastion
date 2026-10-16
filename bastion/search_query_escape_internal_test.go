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
		if r.URL.Path == "/api/"+VersionWallixAPI312+"/usergroups/" && r.URL.Query().Get("q") == "group_name="+name {
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
