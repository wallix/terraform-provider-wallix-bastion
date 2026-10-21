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

// The bodies are the ones of a Bastion 12.4.1.
func TestDeleteDone(t *testing.T) {
	for _, tc := range []struct {
		code int
		body string
		want bool
	}{
		{http.StatusNoContent, "", true},
		{http.StatusOK, "", true},
		{http.StatusNotFound, `{"error":"Resource not found","description":"Device 'dev-id' not found"}`, true},
		{http.StatusNotFound, `{"error":"Resource not found","description":"Domain 'dom-id' not found"}`, true},
		{http.StatusNotFound, `{"error":"Resource not found","description":"The API version \"v9.99\" is unknown."}`, false},
		{http.StatusNotFound, `{"error":"Resource not found","description":"This resource does not exist"}`, false},
		{http.StatusConflict, `{"error":"Constraint violation"}`, false},
	} {
		if got := deleteDone(tc.code, tc.body); got != tc.want {
			t.Errorf("deleteDone(%d, %s) = %v, want %v", tc.code, tc.body, got, tc.want)
		}
	}
}

// A 404 for an unknown API version doesn't say anything about the object, which may still exist:
// every resource that sends a request on delete must return it as an error, not drop the object.
func TestResourcesDeleteKeepsUnknownAPIVersionAsError(t *testing.T) {
	for name, r := range Provider().ResourcesMap {
		t.Run(name, func(t *testing.T) {
			requested := false
			client := newTestAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
				requested = true
				w.Header().Set(contentTypeHeader, contentTypeJSON)
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"error":"Resource not found","description":"The API version \"v3.12\" is unknown."}`)
			})
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})
			d.SetId("still-there")
			diags := r.DeleteContext(t.Context(), d, client)
			if requested && !diags.HasError() {
				t.Error("delete returned no error on a 404 for an unknown API version")
			}
		})
	}
}
