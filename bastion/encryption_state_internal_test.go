package bastion

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func encryptionTestClient(t *testing.T, apiVersion, body string) *Client {
	t.Helper()
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/"+apiVersion+"/encryption") {
			fmt.Fprint(w, body)

			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"Resource not found"}`)
	})
	client.bastionAPIVersion = apiVersion

	return client
}

// Both supported API versions return seal_state (renamed from "encryption" in API 3.8). The
// resource only leaves state when the encryption was never set up: a sealed Bastion, waiting for
// its passphrase after a reboot, keeps it, so that no plan proposes to set the encryption up again.
// The data source tells whether the encryption is set up, and whether the Bastion is sealed.
func TestEncryptionReadsSealState(t *testing.T) {
	for _, apiVersion := range []string{VersionWallixAPI38, VersionWallixAPI312} {
		for _, tc := range []struct {
			sealState   string
			wantInState bool
		}{
			{encryptionNeedSetup, false},
			{encryptionSealed, true},
			{encryptionUnsealed, true},
		} {
			t.Run(apiVersion+" "+tc.sealState, func(t *testing.T) {
				body := `{"seal_state":"` + tc.sealState + `","encryption_mode":"passphrase"}`

				r := resourceEncryption()
				d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"new_passphrase": "x"})
				d.SetId("encryption")
				if diags := r.ReadContext(t.Context(), d, encryptionTestClient(t, apiVersion, body)); diags.HasError() {
					t.Fatalf("resource read: %v", diags)
				}
				if inState := d.Id() != ""; inState != tc.wantInState {
					t.Errorf("resource in state = %v, want %v", inState, tc.wantInState)
				}

				ds := dataSourceEncryption()
				dd := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})
				if diags := ds.ReadContext(t.Context(), dd, encryptionTestClient(t, apiVersion, body)); diags.HasError() {
					t.Fatalf("data source read: %v", diags)
				}
				if got := dd.Get("enabled").(bool); got != tc.wantInState {
					t.Errorf("enabled = %v, want %v", got, tc.wantInState)
				}
				if got := dd.Get("seal_state").(string); got != tc.sealState {
					t.Errorf("seal_state = %q, want %q", got, tc.sealState)
				}
				if got := dd.Get("encryption_mode").(string); got != "passphrase" {
					t.Errorf("encryption_mode = %q, want passphrase", got)
				}
			})
		}
	}
}

// A response without a known seal_state, like the format of the APIs before 3.8, is an error
// instead of a resource silently reported as set up.
func TestEncryptionUnknownStateIsAnError(t *testing.T) {
	client := encryptionTestClient(t, VersionWallixAPI312, `{"encryption":"ready"}`)
	r := resourceEncryption()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"new_passphrase": "x"})
	d.SetId("encryption")
	if diags := r.ReadContext(t.Context(), d, client); !diags.HasError() {
		t.Error("resource read succeeded, want an error for an unknown seal_state")
	}
}
