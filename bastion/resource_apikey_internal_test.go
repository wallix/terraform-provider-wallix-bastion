package bastion

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const testAPIKey = "test-key-returned-once"

// fakeAPIKeys serves an API key endpoint as a Bastion 12.4.3 does: the key is only in the
// X-Auth-Key header of the creation response, GET returns ********.
func fakeAPIKeys(key string, postStatus int, postBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			if postStatus == http.StatusNoContent {
				w.Header().Set("X-Object-Id", "key-id")
				if key != "" {
					w.Header().Set("X-Auth-Key", key)
				}
			}
			w.WriteHeader(postStatus)
			_, _ = w.Write([]byte(postBody))
		case strings.HasSuffix(r.URL.Path, "/"):
			writeJSON(w, []any{})
		default:
			writeJSON(w, map[string]any{
				"id": "key-id", "apikey_name": "ci", "profile": "auditor", "apikey": "********", "ip_limitation": "",
			})
		}
	}
}

func apiKeyResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{"apikey": resourceAPIKey(), "apikey_v2": resourceAPIKeyV2()}
}

func createAPIKey(t *testing.T, r *schema.Resource, handler http.HandlerFunc) (*schema.ResourceData, diag.Diagnostics) {
	t.Helper()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"apikey_name": "ci", skProfile: "auditor"})

	return d, r.CreateContext(t.Context(), d, newTestAPIClient(t, handler))
}

// The provider stored the ******** of GET, so a key created by Terraform couldn't be used.
func TestAPIKeyStoresTheKeyReturnedAtCreation(t *testing.T) {
	for name, r := range apiKeyResources() {
		t.Run(name, func(t *testing.T) {
			handler := fakeAPIKeys(testAPIKey, http.StatusNoContent, "")
			d, diags := createAPIKey(t, r, handler)
			if len(diags) != 0 {
				t.Fatalf("create: %v", diags)
			}
			if got := d.Get("apikey").(string); got != testAPIKey {
				t.Errorf("apikey after create = %q, want the X-Auth-Key header", got)
			}
			if diags := r.ReadContext(t.Context(), d, newTestAPIClient(t, handler)); diags.HasError() {
				t.Fatalf("read: %v", diags)
			}
			if got := d.Get("apikey").(string); got != testAPIKey {
				t.Errorf("apikey after read = %q, want the key stored at creation", got)
			}
		})
	}
}

func TestAPIKeyWarnsWithoutKeyHeader(t *testing.T) {
	for name, r := range apiKeyResources() {
		t.Run(name, func(t *testing.T) {
			d, diags := createAPIKey(t, r, fakeAPIKeys("", http.StatusNoContent, ""))
			if diags.HasError() || len(diags) != 1 || diags[0].Severity != diag.Warning {
				t.Fatalf("create diagnostics = %v, want one warning", diags)
			}
			if got := d.Get("apikey").(string); got != "" {
				t.Errorf("apikey = %q, want empty", got)
			}
		})
	}
}

// The 409 doesn't say which profiles are read-only.
func TestAPIKeyV2ReadOnlyProfileError(t *testing.T) {
	body := `{"error":"Constraint violation","description":"The API key must be associated to a read-only profile"}`
	_, diags := createAPIKey(t, resourceAPIKeyV2(), fakeAPIKeys("", http.StatusConflict, body))
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "editable = false") {
		t.Errorf("create diagnostics = %v, want the read-only profile hint", diags)
	}
}
