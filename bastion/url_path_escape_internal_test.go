package bastion

import (
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// newRecordingClient returns a client of a TLS Bastion API that authenticates any POST /api,
// answers 204 to everything else and records the method and request URI of each other request.
func newRecordingClient(t *testing.T) (*Client, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var requests []string
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.RequestURI)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	return client, func() []string {
		mu.Lock()
		defer mu.Unlock()

		return append([]string(nil), requests...)
	}
}

// A user or timeframe name is a path segment: unescaped, "a#b" was sent as /users/a, which
// deleted (or updated) the user "a" on a Bastion 12.4.
func TestUserAndTimeframeNamesAreEscapedInPaths(t *testing.T) {
	for name, tc := range map[string]struct {
		resource *schema.Resource
		raw      map[string]interface{}
		want     string
	}{
		"user": {
			resource: resourceUser(),
			raw:      map[string]interface{}{"user_name": "a#b"},
			want:     "DELETE /api/v3.12/users/a%23b",
		},
		"timeframe": {
			resource: resourceTimeframe(),
			raw:      map[string]interface{}{"timeframe_name": "a#b"},
			want:     "DELETE /api/v3.12/timeframes/a%23b",
		},
	} {
		t.Run(name, func(t *testing.T) {
			client, requests := newRecordingClient(t)
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.raw)
			d.SetId("a#b")

			if diags := tc.resource.DeleteContext(t.Context(), d, client); diags.HasError() {
				t.Fatalf("delete: %v", diags)
			}
			if got := requests(); len(got) != 1 || got[0] != tc.want {
				t.Errorf("requests = %v, want [%s]", got, tc.want)
			}
		})
	}
}
