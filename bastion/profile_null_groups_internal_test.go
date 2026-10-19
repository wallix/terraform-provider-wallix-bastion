package bastion

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// A limitation enabled without any group reads back with null groups and a null default group
// (reproduced on a Bastion 12.4 with a PUT ?force=true and "target_groups": []). Reading such a
// profile crashed the provider on a refresh, an import and in the data source.
func TestProfileReadWithNullLimitationGroups(t *testing.T) {
	const profile = `{"id":"pid","profile_name":"p","target_access":true,` +
		`"target_groups_limitation":{"enabled":true,"target_groups":null,"default_target_group":null},` +
		`"user_groups_limitation":{"enabled":true,"user_groups":null}}`
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		switch r.URL.Path {
		case "/api/" + VersionWallixAPI312 + "/profiles/":
			fmt.Fprint(w, "["+profile+"]")
		case "/api/" + VersionWallixAPI312 + "/profiles/pid":
			fmt.Fprint(w, profile)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"Resource not found"}`)
		}
	})

	r := resourceProfile()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"profile_name": "p"})
	d.SetId("pid")
	if diags := r.ReadContext(t.Context(), d, client); diags.HasError() {
		t.Fatalf("resource read: %v", diags)
	}
	if got := d.Get("target_groups_limitation.0.default_target_group").(string); got != "" {
		t.Errorf("resource default_target_group = %q, want empty", got)
	}

	ds := dataSourceProfile()
	dd := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"profile_name": "p"})
	if diags := ds.ReadContext(t.Context(), dd, client); diags.HasError() {
		t.Fatalf("data source read: %v", diags)
	}
	if dd.Id() != "pid" {
		t.Errorf("data source id = %q, want pid", dd.Id())
	}
}
