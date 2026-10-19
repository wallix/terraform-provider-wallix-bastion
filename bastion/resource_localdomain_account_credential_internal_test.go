package bastion

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// When the parent of a credential (account, local domain, device or application) is deleted
// outside of Terraform, both the GET of the credential and the list of the credentials of the
// account return 404. The read must drop the credential from state: an error blocked every plan
// until a "terraform state rm" (reproduced on a Bastion 12.4).
func TestLocalDomainAccountCredentialReadDropsCredentialOfDeletedParent(t *testing.T) {
	for name, tc := range map[string]struct {
		resource *schema.Resource
		raw      map[string]interface{}
	}{
		"device": {
			resource: resourceDeviceLocalDomainAccountCredential(),
			raw:      map[string]interface{}{skDeviceID: "dev", skDomainID: "dom", skAccountID: "acc"},
		},
		"application": {
			resource: resourceApplicationLocalDomainAccountCredential(),
			raw:      map[string]interface{}{skApplicationID: "app", skDomainID: "dom", skAccountID: "acc"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			listed := false
			client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/accounts/acc/credentials/") {
					mu.Lock()
					listed = true
					mu.Unlock()
				}
				w.Header().Set(contentTypeHeader, contentTypeJSON)
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"error":"Resource not found"}`)
			})
			tc.raw[skType] = "password"
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.raw)
			d.SetId("cred")

			if diags := tc.resource.ReadContext(t.Context(), d, client); diags.HasError() {
				t.Fatalf("read: %v", diags)
			}
			if d.Id() != "" {
				t.Errorf("id = %q, want the credential dropped from state", d.Id())
			}
			mu.Lock()
			defer mu.Unlock()
			if !listed {
				t.Error("the read didn't look the credential up by account and type")
			}
		})
	}
}
