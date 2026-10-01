package bastion

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testWebApplicationURL         = "https://example.com"
	testWebApplicationLoginURL    = "https://example.com/login"
	testWebApplicationLoginButton = "#submit-button"
)

// webApplicationRaw returns the configuration of a valid web_application, merged with overrides.
func webApplicationRaw(overrides map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"application_name": "web-app",
		skConnectionPolicy: "WEBAPP",
		"category":         "web_application",
		"application_url":  testWebApplicationURL,
	}
	maps.Copy(raw, overrides)

	return raw
}

// webApplicationData returns resource data for the creation of a valid web_application, merged
// with overrides.
func webApplicationData(t *testing.T, overrides map[string]interface{}) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, resourceApplication().Schema, webApplicationRaw(overrides))
}

// webApplicationUpdateData returns resource data for the update of a web_application, from a
// state made of the valid configuration plus stateOverrides, to the valid configuration merged
// with overrides.
func webApplicationUpdateData(
	t *testing.T, stateOverrides map[string]string, overrides map[string]interface{},
) *schema.ResourceData {
	t.Helper()

	attributes := map[string]string{"id": "app-id"}
	for k, v := range webApplicationRaw(nil) {
		attributes[k] = fmt.Sprint(v)
	}
	maps.Copy(attributes, stateOverrides)
	state := &terraform.InstanceState{ID: "app-id", Attributes: attributes}

	sm := schema.InternalMap(resourceApplication().Schema)
	config := terraform.NewResourceConfigRaw(webApplicationRaw(overrides))
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

func TestFillApplicationHandlesNilLocalDomains(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})

	fillApplication(d, jsonApplication{
		ApplicationName:  "test-app",
		ConnectionPolicy: skProtoRDP,
		Category:         skStandard,
	})

	localDomains := d.Get("local_domains")
	if localDomains == nil {
		t.Fatalf("expected local_domains to be initialized")
	}
}

func TestPrepareApplicationJSONRejectsParametersForWebApplication(t *testing.T) {
	d := webApplicationData(t, map[string]interface{}{"parameters": "some-value"})

	_, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err == nil {
		t.Fatalf("expected an error when parameters is set with category = web_application")
	}
	if want := "parameters cannot be configured when category = web_application"; err.Error() != want {
		t.Fatalf("unexpected error message: got %q, want %q", err.Error(), want)
	}
}

func TestPrepareApplicationJSONOmitsParametersForWebApplication(t *testing.T) {
	d := webApplicationData(t, nil)

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Parameters != nil {
		t.Fatalf("expected Parameters to be nil, got %q", *jsonData.Parameters)
	}

	body, err := json.Marshal(jsonData)
	if err != nil {
		t.Fatalf("unexpected error marshaling: %v", err)
	}
	if strings.Contains(string(body), "parameters") {
		t.Fatalf("expected marshaled JSON to omit the parameters key, got: %s", body)
	}
}

// standardApplicationData returns resource data for a valid standard application, merged with overrides.
func standardApplicationData(t *testing.T, overrides map[string]interface{}) *schema.ResourceData {
	t.Helper()

	raw := map[string]interface{}{
		"application_name": "std-app",
		skConnectionPolicy: skProtoRDP,
		"category":         skStandard,
		skTarget:           "cluster",
		"paths": []interface{}{
			map[string]interface{}{
				skTarget:      "Interactive@device:SSH",
				"program":     "application_path",
				"working_dir": "directory",
			},
		},
	}
	maps.Copy(raw, overrides)

	return schema.TestResourceDataRaw(t, resourceApplication().Schema, raw)
}

func TestPrepareApplicationJSONKeepsParametersForStandard(t *testing.T) {
	d := standardApplicationData(t, map[string]interface{}{"parameters": "some-value"})

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Parameters == nil || *jsonData.Parameters != "some-value" {
		t.Fatalf("expected Parameters to be \"some-value\", got %v", jsonData.Parameters)
	}
}

func TestPrepareApplicationJSONUpdateStandard(t *testing.T) {
	d := standardApplicationData(t, nil)

	jsonData, err := prepareApplicationJSON(d, false, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Category != "" {
		t.Fatalf("expected category to be omitted on update, got %q", jsonData.Category)
	}
	if jsonData.Target == nil || *jsonData.Target != "cluster" {
		t.Fatalf("expected target to be sent on update, got %v", jsonData.Target)
	}
	if jsonData.Paths == nil || len(*jsonData.Paths) != 1 {
		t.Fatalf("expected paths to be sent on update, got %v", jsonData.Paths)
	}

	d = standardApplicationData(t, map[string]interface{}{skTarget: ""})
	if _, err = prepareApplicationJSON(d, false, VersionWallixAPI312); err == nil {
		t.Fatalf("expected an error when target is missing on update of a standard application")
	}
}

func TestPrepareApplicationJSONUpdateWebApplication(t *testing.T) {
	d := webApplicationData(t, map[string]interface{}{skGlobalDomains: []interface{}{"domain1"}})

	jsonData, err := prepareApplicationJSON(d, false, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Category != "" {
		t.Fatalf("expected category to be omitted on update, got %q", jsonData.Category)
	}
	if jsonData.ApplicationURL == nil || *jsonData.ApplicationURL != testWebApplicationURL {
		t.Fatalf("expected application_url to be sent on update, got %v", jsonData.ApplicationURL)
	}
	if jsonData.GlobalDomains == nil || len(*jsonData.GlobalDomains) != 1 {
		t.Fatalf("expected global_domains to be sent on update, got %v", jsonData.GlobalDomains)
	}
	if jsonData.Target != nil || jsonData.Paths != nil {
		t.Fatalf("expected target and paths to be omitted for web_application")
	}
}

func TestPrepareApplicationJSONCreateSendsCategory(t *testing.T) {
	d := webApplicationData(t, nil)

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Category != "web_application" {
		t.Fatalf("expected category to be sent on creation, got %q", jsonData.Category)
	}
}

func TestPrepareApplicationJSONSendsWebLoginFields(t *testing.T) {
	d := webApplicationData(t, map[string]interface{}{
		skLoginFormURL:        testWebApplicationLoginURL,
		skLoginButtonSelector: testWebApplicationLoginButton,
		skAllowNonPostForm:    true,
	})

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.LoginFormURL == nil || *jsonData.LoginFormURL != testWebApplicationLoginURL {
		t.Fatalf("expected login_form_url to be sent, got %v", jsonData.LoginFormURL)
	}
	if jsonData.LoginButtonSelector == nil || *jsonData.LoginButtonSelector != testWebApplicationLoginButton {
		t.Fatalf("expected login_button_selector to be sent, got %v", jsonData.LoginButtonSelector)
	}
	if jsonData.AllowNonPostForm == nil || !*jsonData.AllowNonPostForm {
		t.Fatalf("expected allow_non_post_form to be sent as true, got %v", jsonData.AllowNonPostForm)
	}
}

func TestPrepareApplicationJSONOmitsUnsetWebLoginFields(t *testing.T) {
	for _, tc := range []struct {
		name        string
		d           *schema.ResourceData
		newResource bool
	}{
		{"create", webApplicationData(t, nil), true},
		{"unchanged update", webApplicationUpdateData(t, nil, nil), false},
	} {
		jsonData, err := prepareApplicationJSON(tc.d, tc.newResource, VersionWallixAPI312)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		body, err := json.Marshal(jsonData)
		if err != nil {
			t.Fatalf("%s: unexpected error marshaling: %v", tc.name, err)
		}
		for _, key := range []string{skLoginFormURL, skLoginButtonSelector, skAllowNonPostForm} {
			if strings.Contains(string(body), `"`+key+`"`) {
				t.Fatalf("%s: expected marshaled JSON to omit the %s key, got: %s", tc.name, key, body)
			}
		}
	}
}

func TestPrepareApplicationJSONClearsWebLoginFieldsOnUpdate(t *testing.T) {
	d := webApplicationUpdateData(t, map[string]string{
		skLoginFormURL:        testWebApplicationLoginURL,
		skLoginButtonSelector: testWebApplicationLoginButton,
		skAllowNonPostForm:    "true",
	}, nil)

	jsonData, err := prepareApplicationJSON(d, false, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.LoginFormURL == nil || *jsonData.LoginFormURL != "" {
		t.Fatalf("expected login_form_url to be sent empty to clear it, got %v", jsonData.LoginFormURL)
	}
	if jsonData.LoginButtonSelector == nil || *jsonData.LoginButtonSelector != "" {
		t.Fatalf("expected login_button_selector to be sent empty to clear it, got %v", jsonData.LoginButtonSelector)
	}
	if jsonData.AllowNonPostForm == nil || *jsonData.AllowNonPostForm {
		t.Fatalf("expected allow_non_post_form to be sent as false to clear it, got %v", jsonData.AllowNonPostForm)
	}
}

func TestPrepareApplicationJSONRejectsWebLoginFieldsForOtherCategories(t *testing.T) {
	for _, tc := range []struct {
		category   string
		apiVersion string
		key        string
		value      interface{}
	}{
		{skStandard, VersionWallixAPI312, skLoginFormURL, testWebApplicationLoginURL},
		{skStandard, VersionWallixAPI312, skLoginButtonSelector, testWebApplicationLoginButton},
		{skStandard, VersionWallixAPI312, skAllowNonPostForm, true},
		{"jumphost", VersionWallixAPI38, skLoginFormURL, testWebApplicationLoginURL},
		{"jumphost", VersionWallixAPI38, skLoginButtonSelector, testWebApplicationLoginButton},
		{"jumphost", VersionWallixAPI38, skAllowNonPostForm, true},
	} {
		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{
			"application_name": "app",
			skConnectionPolicy: skProtoRDP,
			"category":         tc.category,
			tc.key:             tc.value,
		})

		_, err := prepareApplicationJSON(d, true, tc.apiVersion)
		want := tc.key + " cannot be configured when category = " + tc.category
		if err == nil || err.Error() != want {
			t.Fatalf("expected error %q, got %v", want, err)
		}
	}
}

func TestFillApplicationWebLoginFields(t *testing.T) {
	loginURL, loginButton, allowNonPostForm := testWebApplicationLoginURL, testWebApplicationLoginButton, true
	withFields := jsonApplication{
		ApplicationName:     "web-app",
		ConnectionPolicy:    "WEBAPP",
		Category:            "web_application",
		LoginFormURL:        &loginURL,
		LoginButtonSelector: &loginButton,
		AllowNonPostForm:    &allowNonPostForm,
	}
	withoutFields := withFields
	withoutFields.LoginFormURL, withoutFields.LoginButtonSelector, withoutFields.AllowNonPostForm = nil, nil, nil

	for name, tc := range map[string]struct {
		resourceSchema map[string]*schema.Schema
		fill           func(*schema.ResourceData, jsonApplication)
	}{
		"resource":    {resourceApplication().Schema, fillApplication},
		"data source": {dataSourceApplication().Schema, fillSourceApplication},
	} {
		d := schema.TestResourceDataRaw(t, tc.resourceSchema, map[string]interface{}{})

		tc.fill(d, withFields)
		if got := d.Get(skLoginFormURL).(string); got != loginURL {
			t.Fatalf("%s: expected login_form_url %q, got %q", name, loginURL, got)
		}
		if got := d.Get(skLoginButtonSelector).(string); got != loginButton {
			t.Fatalf("%s: expected login_button_selector %q, got %q", name, loginButton, got)
		}
		if !d.Get(skAllowNonPostForm).(bool) {
			t.Fatalf("%s: expected allow_non_post_form to be true", name)
		}

		tc.fill(d, withoutFields)
		if d.Get(skLoginFormURL).(string) != "" || d.Get(skLoginButtonSelector).(string) != "" ||
			d.Get(skAllowNonPostForm).(bool) {
			t.Fatalf("%s: expected login automation fields to be reset when absent from the API response", name)
		}
	}
}
