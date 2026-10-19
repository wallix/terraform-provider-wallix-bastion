package bastion

import (
	"fmt"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func authorizationRaw(overrides map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"authorization_name": "authorization",
		"user_group":         "users",
		"target_group":       "targets",
		"authorize_sessions": true,
	}
	maps.Copy(raw, overrides)

	return raw
}

// authorizationUpdateData returns resource data for the update of an authorization, from a
// state made of the base configuration plus stateOverrides, to the base configuration merged
// with overrides.
func authorizationUpdateData(
	t *testing.T, stateOverrides map[string]string, overrides map[string]interface{},
) *schema.ResourceData {
	t.Helper()

	attributes := map[string]string{"id": "authorization-id"}
	for k, v := range authorizationRaw(nil) {
		attributes[k] = fmt.Sprint(v)
	}
	maps.Copy(attributes, stateOverrides)
	state := &terraform.InstanceState{ID: "authorization-id", Attributes: attributes}

	sm := schema.InternalMap(resourceAuthorization().Schema)
	config := terraform.NewResourceConfigRaw(authorizationRaw(overrides))
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

func TestPrepareAuthorizationJSONSessionSharingOnCreate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAuthorization().Schema, authorizationRaw(map[string]interface{}{
		"authorize_session_sharing": true,
		"session_sharing_mode":      "view_only",
	}))
	jsonData := prepareAuthorizationJSON(d, true)
	if jsonData.AuthorizeSessionSharing == nil || !*jsonData.AuthorizeSessionSharing ||
		jsonData.SessionSharingMode != "view_only" {
		t.Errorf("authorize_session_sharing, mode = %v, %q, want true, view_only",
			jsonData.AuthorizeSessionSharing, jsonData.SessionSharingMode)
	}

	// Without sharing, nothing is sent, as before.
	d = schema.TestResourceDataRaw(t, resourceAuthorization().Schema, authorizationRaw(nil))
	if jsonData := prepareAuthorizationJSON(d, true); jsonData.AuthorizeSessionSharing != nil ||
		jsonData.SessionSharingMode != "" {
		t.Errorf("authorize_session_sharing, mode = %v, %q, want them omitted",
			*jsonData.AuthorizeSessionSharing, jsonData.SessionSharingMode)
	}
}

// Disabling sharing was never sent (the field was omitted when false), so it stayed enabled on
// the Bastion.
func TestPrepareAuthorizationJSONDisablesSessionSharing(t *testing.T) {
	d := authorizationUpdateData(t,
		map[string]string{"authorize_session_sharing": "true", "session_sharing_mode": "view_only"},
		map[string]interface{}{})
	jsonData := prepareAuthorizationJSON(d, false)
	if jsonData.AuthorizeSessionSharing == nil || *jsonData.AuthorizeSessionSharing {
		t.Errorf("authorize_session_sharing = %v, want false sent", jsonData.AuthorizeSessionSharing)
	}
	if jsonData.SessionSharingMode != "" {
		t.Errorf("session_sharing_mode = %q, want it omitted when disabling", jsonData.SessionSharingMode)
	}
}

// An update that doesn't touch sharing sends the same request as before (034a630), also for a
// state written before these attributes existed.
func TestPrepareAuthorizationJSONUnchangedSharingIsOmitted(t *testing.T) {
	for name, state := range map[string]map[string]string{
		"disabled":                {"authorize_session_sharing": "false", skDescription: "first"},
		"state without attribute": {skDescription: "first"},
	} {
		t.Run(name, func(t *testing.T) {
			d := authorizationUpdateData(t, state, map[string]interface{}{skDescription: "second"})
			if jsonData := prepareAuthorizationJSON(d, false); jsonData.AuthorizeSessionSharing != nil {
				t.Errorf("authorize_session_sharing = %v, want it omitted", *jsonData.AuthorizeSessionSharing)
			}
		})
	}
}
