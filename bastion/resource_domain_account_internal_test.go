package bastion

import (
	"maps"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func domainAccountRaw(overrides map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		skDomainID:     "domain-id",
		skAccountName:  "account",
		skAccountLogin: "login",
	}
	maps.Copy(raw, overrides)

	return raw
}

// domainAccountUpdateData returns resource data for the update of a domain account whose state
// has the resource "device:SSH", to the base configuration merged with overrides.
func domainAccountUpdateData(t *testing.T, overrides map[string]interface{}) *schema.ResourceData {
	t.Helper()

	resources := resourceDomainAccount().Schema["resources"]
	hash := schema.HashSchema(resources.Elem.(*schema.Schema))("device:SSH")
	attributes := map[string]string{
		"id":                              "account-id",
		skDomainID:                        "domain-id",
		skAccountName:                     "account",
		skAccountLogin:                    "login",
		"resources.#":                     "1",
		"resources." + strconv.Itoa(hash): "device:SSH",
	}
	state := &terraform.InstanceState{ID: "account-id", Attributes: attributes}

	sm := schema.InternalMap(resourceDomainAccount().Schema)
	config := terraform.NewResourceConfigRaw(domainAccountRaw(overrides))
	diff, err := sm.Diff(t.Context(), state, config, nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error computing the diff: %v", err)
	}
	d, err := sm.Data(state, diff)
	if err != nil {
		t.Fatalf("unexpected error building the resource data: %v", err)
	}

	return d
}

// Emptying resources was never sent (the field was omitted when empty), so the account kept its
// access to the resources.
func TestPrepareDomainAccountJSONSendsEmptiedResources(t *testing.T) {
	d := domainAccountUpdateData(t, map[string]interface{}{"resources": []interface{}{}})
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Resources == nil || len(*jsonData.Resources) != 0 {
		t.Errorf("resources = %v, want an empty list sent", jsonData.Resources)
	}
}

// Without resources in the configuration (computed), the resources in state are sent unchanged
// instead of being removed (718a522).
func TestPrepareDomainAccountJSONKeepsUnsetResources(t *testing.T) {
	d := domainAccountUpdateData(t, map[string]interface{}{skDescription: "changed"})
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Resources == nil || len(*jsonData.Resources) != 1 || (*jsonData.Resources)[0] != "device:SSH" {
		t.Errorf("resources = %v, want the resources in state, [device:SSH]", jsonData.Resources)
	}
}

func TestPrepareDomainAccountJSONSendsResources(t *testing.T) {
	d := domainAccountUpdateData(t, map[string]interface{}{"resources": []interface{}{"device:RDP"}})
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Resources == nil || len(*jsonData.Resources) != 1 || (*jsonData.Resources)[0] != "device:RDP" {
		t.Errorf("resources = %v, want [device:RDP]", jsonData.Resources)
	}
}
