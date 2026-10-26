package bastion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// A password_length with no character class failed only at apply (400 from the API).
func TestPasswordChangePolicyCharClassPlan(t *testing.T) {
	const (
		length  = "password_length"
		unknown = "74D93920-ED26-11E3-AC10-0800200C9A66" // the SDK's marker of an unknown value
	)
	r := resourcePasswordChangePolicy()
	for name, tc := range map[string]struct {
		id      string
		config  map[string]interface{}
		wantErr bool
	}{
		"length without class":      {config: map[string]interface{}{length: 14}, wantErr: true},
		"length with class at 0":    {config: map[string]interface{}{length: 14, "upper_chars": 0}},
		"length with unknown class": {config: map[string]interface{}{length: 14, "lower_chars": unknown}},
		"unknown length":            {config: map[string]interface{}{length: unknown}},
		"ssh key without length":    {config: map[string]interface{}{"ssh_key_type": "ED25519"}},
		"update without class":      {id: "pcp-plan-id", config: map[string]interface{}{length: 14}},
	} {
		t.Run(name, func(t *testing.T) {
			config := map[string]interface{}{"password_change_policy_name": "pcp-plan"}
			attrs := map[string]string{"password_change_policy_name": "pcp-plan"}
			for k, v := range tc.config {
				config[k] = v
				attrs[k] = fmt.Sprint(v)
			}
			// r.Diff doesn't build the raw config that Terraform sends in a real plan.
			rawConfig, err := (&terraform.InstanceState{Attributes: attrs}).AttrsAsObjectValue(
				r.CoreConfigSchema().ImpliedType())
			if err != nil {
				t.Fatalf("building the raw config: %v", err)
			}
			state := &terraform.InstanceState{ID: tc.id, RawConfig: rawConfig}
			if tc.id != "" {
				state.Attributes = map[string]string{
					"id": tc.id, "password_change_policy_name": "pcp-plan", length: "12", "digit_chars": "1",
				}
			}
			_, err = r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), nil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("plan error = %v, want an error: %t", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "At least one number of chars") {
				t.Errorf("plan error = %v, want the API rule", err)
			}
		})
	}
}
