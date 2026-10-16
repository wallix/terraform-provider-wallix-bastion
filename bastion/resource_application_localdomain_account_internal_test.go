package bastion

import (
	"fmt"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const appAccountPassword = "secret"

func appAccountRaw(overrides map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		skApplicationID: "app-id",
		skDomainID:      "domain-id",
		skAccountName:   "account",
		skAccountLogin:  "login",
	}
	maps.Copy(raw, overrides)

	return raw
}

// appAccountUpdateData returns resource data for the update of an account, from a state made of
// the base configuration plus stateOverrides, to the base configuration merged with overrides.
func appAccountUpdateData(
	t *testing.T, stateOverrides map[string]string, overrides map[string]interface{},
) *schema.ResourceData {
	t.Helper()

	attributes := map[string]string{"id": "account-id"}
	for k, v := range appAccountRaw(nil) {
		attributes[k] = fmt.Sprint(v)
	}
	maps.Copy(attributes, stateOverrides)
	state := &terraform.InstanceState{ID: "account-id", Attributes: attributes}

	sm := schema.InternalMap(resourceApplicationLocalDomainAccount().Schema)
	config := terraform.NewResourceConfigRaw(appAccountRaw(overrides))
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

func TestPrepareApplicationLocalDomainAccountJSONCreate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceApplicationLocalDomainAccount().Schema,
		appAccountRaw(map[string]interface{}{skPassword: appAccountPassword}))
	jsonData := prepareApplicationLocalDomainAccountJSON(d, true)
	if jsonData.Credentials == nil || len(*jsonData.Credentials) != 1 ||
		(*jsonData.Credentials)[0].Password != appAccountPassword {
		t.Errorf("credentials = %v, want the password on creation", jsonData.Credentials)
	}

	d = schema.TestResourceDataRaw(t, resourceApplicationLocalDomainAccount().Schema, appAccountRaw(nil))
	if jsonData := prepareApplicationLocalDomainAccountJSON(d, true); jsonData.Credentials != nil {
		t.Errorf("credentials = %v, want none without password", *jsonData.Credentials)
	}
}

// The update is sent with force=true: sending credentials (even empty) replaces the credentials
// of the account on the Bastion, which wiped the password on every unrelated change.
func TestPrepareApplicationLocalDomainAccountJSONUpdateKeepsCredentials(t *testing.T) {
	for name, tc := range map[string]struct {
		state  map[string]string
		config map[string]interface{}
	}{
		"no password, description changed": {
			state:  map[string]string{skDescription: "first"},
			config: map[string]interface{}{skDescription: "second"},
		},
		"same password, description changed": {
			state:  map[string]string{skDescription: "first", skPassword: appAccountPassword},
			config: map[string]interface{}{skDescription: "second", skPassword: appAccountPassword},
		},
		"password removed from the configuration": {
			state:  map[string]string{skPassword: appAccountPassword},
			config: map[string]interface{}{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			d := appAccountUpdateData(t, tc.state, tc.config)
			if jsonData := prepareApplicationLocalDomainAccountJSON(d, false); jsonData.Credentials != nil {
				t.Errorf("credentials = %v, want none so that the update keeps them", *jsonData.Credentials)
			}
		})
	}
}

func TestPrepareApplicationLocalDomainAccountJSONUpdateSendsNewPassword(t *testing.T) {
	d := appAccountUpdateData(t,
		map[string]string{skPassword: "old"}, map[string]interface{}{skPassword: "new"})
	jsonData := prepareApplicationLocalDomainAccountJSON(d, false)
	if jsonData.Credentials == nil || len(*jsonData.Credentials) != 1 ||
		(*jsonData.Credentials)[0].Password != "new" {
		t.Errorf("credentials = %v, want the new password", jsonData.Credentials)
	}
}
