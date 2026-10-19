package bastion

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// An object deleted outside of Terraform answers 404 to its DELETE: the destroy must treat it as
// done, not fail and leave the objects depending on it. This happens when the state wasn't
// refreshed (-refresh=false, a saved plan applied later) or when the Bastion deleted it along
// with its parent.
func TestResourcesDeleteTreatsNotFoundAsDeleted(t *testing.T) {
	client := newTestAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"Resource not found"}`)
	})
	for name, r := range Provider().ResourcesMap {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})
			d.SetId("gone")
			if diags := r.DeleteContext(t.Context(), d, client); diags.HasError() {
				t.Errorf("delete: %v", diags)
			}
		})
	}
}
