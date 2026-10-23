package bastion

import (
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// recordRequests returns a handler that answers 204 and records the method and request URI of
// each request.
func recordRequests() (http.HandlerFunc, func() []string) {
	var mu sync.Mutex
	var requests []string
	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.RequestURI)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}

	return handler, func() []string {
		mu.Lock()
		defer mu.Unlock()

		return append([]string(nil), requests...)
	}
}

// A user or timeframe name is a path segment: unescaped, "a#b" was sent as /users/a, which
// deleted (or updated) the user "a" on a Bastion 12.4.
func TestUserAndTimeframeNamesAreEscapedInPaths(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		handler, requests := recordRequests()
		client := newTestAPIClient(t, handler)
		d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{"user_name": "a#b"})
		d.SetId("a#b")

		if diags := resourceUser().DeleteContext(t.Context(), d, client); diags.HasError() {
			t.Fatalf("delete: %v", diags)
		}
		if got, want := requests(), "DELETE /api/v3.12/users/a%23b"; len(got) != 1 || got[0] != want {
			t.Errorf("requests = %v, want [%s]", got, want)
		}
	})
	t.Run("timeframe", func(t *testing.T) {
		handler, requests := recordRequests()
		server := newTestProviderServer(t, newTestAPIClient(t, handler))
		typ := schemaType(t, server, timeframeType)

		resp, err := server.ApplyResourceChange(t.Context(), &tfprotov5.ApplyResourceChangeRequest{
			TypeName:     timeframeType,
			PriorState:   dynamicValue(t, typ, timeframeValues("a#b", nil)),
			PlannedState: nullValue(t, typ),
			Config:       nullValue(t, typ),
		})
		if err != nil || hasError(resp.Diagnostics) {
			t.Fatalf("delete: %v %v", err, resp.Diagnostics)
		}
		if got, want := requests(), "DELETE /api/v3.12/timeframes/a%23b"; len(got) != 1 || got[0] != want {
			t.Errorf("requests = %v, want [%s]", got, want)
		}
	})
}
