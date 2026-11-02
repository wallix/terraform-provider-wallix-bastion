package bastion

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testProfileName         = "profile_name"
	testProfileGuiFeatures  = "gui_features"
	testProfileTransmission = "gui_transmission"
	testProfileType         = "wallix-bastion_profile"
	testProfileDescription  = "changed"
)

// editableState is the editable of a profile state: unset (API v3.8), true or false.
type editableState int

const (
	editableUnset editableState = iota
	editableTrue
	editableFalse
)

func profileAttributes(editable editableState) map[string]string {
	attrs := map[string]string{
		"id": "p-id", testProfileName: "ro", testProfileGuiFeatures + ".#": "1",
		testProfileGuiFeatures + ".0.wab_audit": skView, testProfileTransmission + ".#": "1",
	}
	if editable != editableUnset {
		attrs[skEditable] = strconv.FormatBool(editable == editableTrue)
	}

	return attrs
}

// objectOf returns the raw value of a resource with these flat attributes, which r.Diff doesn't build.
func objectOf(t *testing.T, r *schema.Resource, attrs map[string]string) *terraform.InstanceState {
	t.Helper()
	raw, err := (&terraform.InstanceState{Attributes: attrs}).AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
	if err != nil {
		t.Fatalf("building the raw value: %v", err)
	}

	return &terraform.InstanceState{ID: attrs["id"], Attributes: attrs, RawState: raw, RawConfig: raw}
}

func profileConfig(extra map[string]interface{}) *terraform.ResourceConfig {
	config := map[string]interface{}{
		testProfileName:         "ro",
		testProfileGuiFeatures:  []interface{}{map[string]interface{}{"wab_audit": skView}},
		testProfileTransmission: []interface{}{map[string]interface{}{}},
	}
	for k, v := range extra {
		config[k] = v
	}

	return terraform.NewResourceConfigRaw(config)
}

func block(attrs map[string]interface{}) []interface{} {
	return []interface{}{attrs}
}

// The Bastion refuses any change to a read-only profile (409 on a Bastion 12.4.3), so it is replaced.
func TestProfileReadOnlyChangeReplaces(t *testing.T) {
	r := resourceProfile()
	description := map[string]interface{}{skDescription: testProfileDescription}
	for name, tc := range map[string]struct {
		state       editableState
		config      map[string]interface{}
		wantReplace bool
	}{
		"read-only, description changed": {editableFalse, description, true},
		"read-only, features changed": {editableFalse, map[string]interface{}{
			testProfileGuiFeatures: block(map[string]interface{}{skUsers: skView}),
		}, true},
		"read-only, made editable":    {editableFalse, map[string]interface{}{skEditable: true}, true},
		"read-only, dashboards added": {editableFalse, map[string]interface{}{"dashboards": []interface{}{"audit"}}, true},
		"read-only, limitation added": {editableFalse, map[string]interface{}{
			"user_groups_limitation": block(map[string]interface{}{skUserGroups: []interface{}{"g"}}),
		}, true},
		"editable, description changed":   {editableTrue, description, false},
		"editable, made read-only":        {editableTrue, map[string]interface{}{skEditable: false}, false},
		"no editable (API v3.8), changed": {editableUnset, description, false},
	} {
		t.Run(name, func(t *testing.T) {
			diff, err := r.Diff(t.Context(), objectOf(t, r, profileAttributes(tc.state)), profileConfig(tc.config), nil)
			if err != nil {
				t.Fatalf("plan: %v", err)
			}
			if diff == nil || diff.Empty() {
				t.Fatal("empty plan, want a change")
			}
			if got := diff.RequiresNew(); got != tc.wantReplace {
				t.Errorf("replace = %t, want %t", got, tc.wantReplace)
			}
		})
	}
}

// editable is only sent when set (on creation) or changed (on update); API v3.8 refuses it (400).
func TestProfileEditableRequestBody(t *testing.T) {
	r := resourceProfile()
	for name, tc := range map[string]struct {
		state    editableState // unset: creation
		editable *bool
		version  string
		want     *bool
		wantErr  bool
	}{
		"create, unset":           {version: VersionWallixAPI312},
		"create, false":           {editable: new(false), version: VersionWallixAPI312, want: new(false)},
		"create, false, API v3.8": {editable: new(false), version: "v3.8", wantErr: true},
		"update, made read-only": {
			state: editableTrue, editable: new(false), version: VersionWallixAPI312, want: new(false),
		},
		"update, editable unchanged": {state: editableTrue, editable: new(true), version: VersionWallixAPI312},
	} {
		t.Run(name, func(t *testing.T) {
			var state *terraform.InstanceState
			if tc.state != editableUnset {
				state = objectOf(t, r, profileAttributes(tc.state))
			}
			config := map[string]interface{}{skDescription: testProfileDescription}
			configAttrs := profileAttributes(editableUnset)
			configAttrs[skDescription] = testProfileDescription
			if tc.editable != nil {
				config[skEditable] = *tc.editable
				configAttrs[skEditable] = strconv.FormatBool(*tc.editable)
			}
			diff, err := r.Diff(t.Context(), state, profileConfig(config), nil)
			if err != nil {
				t.Fatalf("plan: %v", err)
			}
			diff.RawConfig = objectOf(t, r, configAttrs).RawConfig
			d, err := schema.InternalMap(r.Schema).Data(state, diff)
			if err != nil {
				t.Fatalf("building the resource data: %v", err)
			}
			got, err := profileEditable(d, tc.version, state == nil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want an error: %t", err, tc.wantErr)
			}
			if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
				t.Errorf("editable sent = %v, want %v", got, tc.want)
			}
		})
	}
}

// API v3.8 doesn't return editable (checked on a Bastion 12.4.3): the data source leaves it null.
func TestProfileDataSourceEditable(t *testing.T) {
	for name, tc := range map[string]struct {
		editable any
		want     tftypes.Value
	}{
		"API v3.12, built-in": {false, tftypes.NewValue(tftypes.Bool, false)},
		"API v3.8":            {nil, tftypes.NewValue(tftypes.Bool, nil)},
	} {
		t.Run(name, func(t *testing.T) {
			profile := map[string]any{"id": "p-id", testProfileName: "auditor"}
			if tc.editable != nil {
				profile[skEditable] = tc.editable
			}
			server := newTestProviderServer(t, newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/") {
					writeJSON(w, []any{profile})

					return
				}
				writeJSON(w, profile)
			}))
			typ := schemaType(t, server, "data."+testProfileType)
			resp, err := server.ReadDataSource(t.Context(), &tfprotov5.ReadDataSourceRequest{
				TypeName: testProfileType,
				Config:   dynamicValue(t, typ, map[string]interface{}{testProfileName: "auditor"}),
			})
			if err != nil || hasError(resp.Diagnostics) {
				t.Fatalf("read: %v %v", err, resp.Diagnostics)
			}
			if got := stateAttributes(t, resp.State, typ)[skEditable]; !got.Equal(tc.want) {
				t.Errorf("editable = %v, want %v", got, tc.want)
			}
		})
	}
}
