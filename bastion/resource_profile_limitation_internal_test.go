package bastion

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// A limitation without a default target group, set from the GUI or the API, is returned with
// "default_target_group": null: the read dereferenced it and crashed the provider.
func TestProfileReadOfALimitationWithNulls(t *testing.T) {
	const body = `{"id":"p","profile_name":"p",` +
		`"target_groups_limitation":{"enabled":true,"default_target_group":null,"target_groups":null},` +
		`"user_groups_limitation":{"enabled":true,"user_groups":null}}`
	var jsonData jsonProfile
	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		t.Fatal(err)
	}
	for name, fill := range map[string]struct {
		resource *schema.Resource
		fill     func(*schema.ResourceData, jsonProfile)
	}{
		"resource":    {resourceProfile(), fillProfile},
		"data source": {dataSourceProfile(), fillSourceProfile},
	} {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, fill.resource.Schema, map[string]interface{}{})
			fill.fill(d, jsonData)
			if got := d.Get("target_groups_limitation.0.default_target_group").(string); got != "" {
				t.Errorf("default_target_group = %q, want empty", got)
			}
			if got := d.Get("target_groups_limitation.#").(int); got != 1 {
				t.Errorf("target_groups_limitation count = %d, want 1", got)
			}
		})
	}
}
