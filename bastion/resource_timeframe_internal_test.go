package bastion

import (
	"fmt"
	"net/http"
	"path"
	"slices"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The API renames a timeframe on PUT; the provider kept the old name as id, so the read after
// the update didn't find it and dropped it from state ("Root object was present, but now absent").
func TestTimeframeRenameFollowsTheNewName(t *testing.T) {
	var mu sync.Mutex
	var requests []string
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		mu.Unlock()
		if r.Method == http.MethodGet {
			w.Header().Set(contentTypeHeader, contentTypeJSON)
			fmt.Fprintf(w, `{"timeframe_name":%q,"description":"","is_overtimable":false,"periods":[]}`,
				path.Base(r.URL.Path))

			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	state := &terraform.InstanceState{ID: "old", Attributes: map[string]string{
		"id":             "old",
		"timeframe_name": "old",
	}}
	sm := schema.InternalMap(resourceTimeframe().Schema)
	diff, err := sm.Diff(t.Context(), state,
		terraform.NewResourceConfigRaw(map[string]interface{}{"timeframe_name": "new"}), nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error computing the diff: %v", err)
	}
	d, err := sm.Data(state, diff)
	if err != nil {
		t.Fatalf("unexpected error building the resource data: %v", err)
	}

	if diags := resourceTimeframeUpdate(t.Context(), d, client); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}
	if d.Id() != "new" {
		t.Errorf("id = %q after the rename, want new", d.Id())
	}
	const base = "/api/" + VersionWallixAPI312
	for _, want := range []string{"PUT " + base + "/timeframes/old", "GET " + base + "/timeframes/new"} {
		if !slices.Contains(requests, want) {
			t.Errorf("requests = %v, want %q", requests, want)
		}
	}
}
