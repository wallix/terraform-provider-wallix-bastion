package bastion

import (
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// A lock with max_duration != duration and no extension failed only at apply (400 from the API).
func TestCheckoutPolicyMaxDurationPlan(t *testing.T) {
	const (
		dur, maxDur, ext = "duration", "max_duration", "extension"
		unknown          = "74D93920-ED26-11E3-AC10-0800200C9A66" // the SDK's marker of an unknown value
	)
	type cfg = map[string]interface{}
	r := resourceCheckoutPolicy()
	for name, tc := range map[string]struct {
		id      string
		config  cfg
		wantErr bool
	}{
		"lock without extension": {config: cfg{skEnableLock: true, dur: 3600, maxDur: 7200}, wantErr: true},
		"lock with extension 0":  {config: cfg{skEnableLock: true, dur: 3600, maxDur: 7200, ext: 0}, wantErr: true},
		"lock with equal max":    {config: cfg{skEnableLock: true, dur: 3600, maxDur: 3600}},
		"lock with extension":    {config: cfg{skEnableLock: true, dur: 3600, maxDur: 7200, ext: 600}},
		"no lock":                {config: cfg{skEnableLock: false, dur: 3600, maxDur: 7200}},
		"lock with unknown max":  {config: cfg{skEnableLock: true, dur: 3600, maxDur: unknown}},
		"update without extension": {
			id: "co-plan-id", config: cfg{skEnableLock: true, dur: 3600, maxDur: 7200}, wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			tc.config["checkout_policy_name"] = "co-plan"
			state := &terraform.InstanceState{ID: tc.id}
			if tc.id != "" {
				state.Attributes = map[string]string{
					"id": tc.id, "checkout_policy_name": "co-plan", skEnableLock: strconv.FormatBool(true),
					dur: "3600", maxDur: "3600",
				}
			}
			_, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(tc.config), nil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("plan error = %v, want an error: %t", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "must be the same as duration") {
				t.Errorf("plan error = %v, want the API rule", err)
			}
		})
	}
}
