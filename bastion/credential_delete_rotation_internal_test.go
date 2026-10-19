package bastion

import (
	"fmt"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// A rotation recreates a credential under a new id, which the read only follows after a refresh.
// Destroyed without one (-refresh=false, a saved plan), the DELETE on the stored id answers 404:
// the delete must remove the credential of the same account and type, or the rotated credential
// stays on the Bastion out of state. When the account has no such credential, the delete is done.
func TestCredentialDeleteFollowsRotation(t *testing.T) {
	const domainID, accountID = "dom", "acc"
	variants := map[string]struct {
		resource    *schema.Resource
		raw         map[string]interface{}
		credentials string
	}{
		"domain": {
			resource:    resourceDomainAccountCredential(),
			raw:         map[string]interface{}{skDomainID: domainID, skAccountID: accountID},
			credentials: "/domains/dom/accounts/acc/credentials/",
		},
		"device": {
			resource:    resourceDeviceLocalDomainAccountCredential(),
			raw:         map[string]interface{}{skDeviceID: "dev", skDomainID: domainID, skAccountID: accountID},
			credentials: "/devices/dev/localdomains/dom/accounts/acc/credentials/",
		},
		"application": {
			resource:    resourceApplicationLocalDomainAccountCredential(),
			raw:         map[string]interface{}{skApplicationID: "app", skDomainID: domainID, skAccountID: accountID},
			credentials: "/applications/app/localdomains/dom/accounts/acc/credentials/",
		},
	}
	for name, v := range variants {
		for listBody, wantDeletes := range map[string][]string{
			`[{"id":"rotated","type":"password"}]`: {"stored", "rotated"},
			`[]`:                                   {"stored"},
		} {
			t.Run(name+" "+listBody, func(t *testing.T) {
				credentials := "/api/" + VersionWallixAPI312 + v.credentials
				var mu sync.Mutex
				var deletes []string
				client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set(contentTypeHeader, contentTypeJSON)
					switch {
					case r.Method == http.MethodGet && r.URL.Path == credentials:
						fmt.Fprint(w, listBody)
					case r.Method == http.MethodDelete && r.URL.Path == credentials+"rotated":
						mu.Lock()
						deletes = append(deletes, "rotated")
						mu.Unlock()
						w.WriteHeader(http.StatusNoContent)
					case r.Method == http.MethodDelete && r.URL.Path == credentials+"stored":
						mu.Lock()
						deletes = append(deletes, "stored")
						mu.Unlock()
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"error":"Resource not found"}`)
					default:
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"error":"Resource not found"}`)
					}
				})
				raw := map[string]interface{}{skType: "password"}
				for k, val := range v.raw {
					raw[k] = val
				}
				d := schema.TestResourceDataRaw(t, v.resource.Schema, raw)
				d.SetId("stored")

				if diags := v.resource.DeleteContext(t.Context(), d, client); diags.HasError() {
					t.Fatalf("delete: %v", diags)
				}
				mu.Lock()
				defer mu.Unlock()
				if !slices.Equal(deletes, wantDeletes) {
					t.Errorf("DELETE requests on %v, want %v", deletes, wantDeletes)
				}
			})
		}
	}
}
