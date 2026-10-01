package bastion

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	testAppName        = "application_name"
	testAppCategory    = "category"
	testAppURLKey      = "application_url"
	testAppCategoryWeb = "web_application"
	testAppURL         = "https://example.com"
)

// testWebApplicationResourceData builds a web_application resource whose fields can be extended or overridden.
func testWebApplicationResourceData(t *testing.T, extra map[string]interface{}) *schema.ResourceData {
	t.Helper()

	raw := map[string]interface{}{
		testAppName:        "web-app",
		skConnectionPolicy: "WEBAPP",
		testAppCategory:    testAppCategoryWeb,
		testAppURLKey:      testAppURL,
	}
	for k, v := range extra {
		raw[k] = v
	}

	return schema.TestResourceDataRaw(t, resourceApplication().Schema, raw)
}

func testStandardApplicationResourceData(t *testing.T, extra map[string]interface{}) *schema.ResourceData {
	t.Helper()

	raw := map[string]interface{}{
		testAppName:        "std-app",
		skConnectionPolicy: "RDP",
		testAppCategory:    skStandard,
		skTarget:           "cluster",
		"paths": []interface{}{
			map[string]interface{}{
				skTarget:      "Interactive@device:SSH",
				"program":     "application_path",
				"working_dir": "directory",
			},
		},
	}
	for k, v := range extra {
		raw[k] = v
	}

	return schema.TestResourceDataRaw(t, resourceApplication().Schema, raw)
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
	d := testWebApplicationResourceData(t, map[string]interface{}{"parameters": "some-value"})

	_, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err == nil {
		t.Fatalf("expected an error when parameters is set with category = web_application")
	}
	if want := "parameters cannot be configured when category = web_application"; err.Error() != want {
		t.Fatalf("unexpected error message: got %q, want %q", err.Error(), want)
	}
}

func TestPrepareApplicationJSONOmitsParametersForWebApplication(t *testing.T) {
	d := testWebApplicationResourceData(t, nil)

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

func TestPrepareApplicationJSONKeepsParametersForStandard(t *testing.T) {
	d := testStandardApplicationResourceData(t, map[string]interface{}{"parameters": "some-value"})

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Parameters == nil || *jsonData.Parameters != "some-value" {
		t.Fatalf("expected Parameters to be \"some-value\", got %v", jsonData.Parameters)
	}
}

func TestPrepareApplicationJSONUpdateWebApplication(t *testing.T) {
	d := testWebApplicationResourceData(t, map[string]interface{}{skGlobalDomains: []interface{}{"domain1"}})

	jsonData, err := prepareApplicationJSON(d, false, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Category != "" {
		t.Fatalf("expected category to be omitted on update, got %q", jsonData.Category)
	}
	if jsonData.ApplicationURL == nil || *jsonData.ApplicationURL != testAppURL {
		t.Fatalf("expected application_url to be sent on update, got %v", jsonData.ApplicationURL)
	}
	if jsonData.GlobalDomains == nil || len(*jsonData.GlobalDomains) != 1 {
		t.Fatalf("expected global_domains to be sent on update, got %v", jsonData.GlobalDomains)
	}
	if jsonData.Target != nil || jsonData.Paths != nil {
		t.Fatalf("expected target and paths to be omitted for web_application")
	}
}

func TestPrepareApplicationJSONUpdateStandard(t *testing.T) {
	d := testStandardApplicationResourceData(t, nil)

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

	d = testStandardApplicationResourceData(t, map[string]interface{}{skTarget: ""})
	if _, err = prepareApplicationJSON(d, false, VersionWallixAPI312); err == nil {
		t.Fatalf("expected an error when target is missing on update of a standard application")
	}
}

func TestPrepareApplicationJSONCreateSendsCategory(t *testing.T) {
	d := testWebApplicationResourceData(t, nil)

	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Category != testAppCategoryWeb {
		t.Fatalf("expected category to be sent on creation, got %q", jsonData.Category)
	}
}

func TestPrepareApplicationJSONAllowNonPostForm(t *testing.T) {
	for _, want := range []bool{false, true} {
		d := testWebApplicationResourceData(t, map[string]interface{}{"allow_non_post_form": want})

		jsonData, err := prepareApplicationJSON(d, false, VersionWallixAPI312)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if jsonData.AllowNonPostForm == nil || *jsonData.AllowNonPostForm != want {
			t.Fatalf("expected allow_non_post_form to be sent as %t, got %v", want, jsonData.AllowNonPostForm)
		}
	}
}

func TestPrepareApplicationJSONRejectsAllowNonPostFormForStandard(t *testing.T) {
	d := testStandardApplicationResourceData(t, map[string]interface{}{"allow_non_post_form": true})

	_, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err == nil {
		t.Fatalf("expected an error when allow_non_post_form is set with category = standard")
	}
	if want := "allow_non_post_form cannot be configured when category = standard"; err.Error() != want {
		t.Fatalf("unexpected error message: got %q, want %q", err.Error(), want)
	}

	d = testStandardApplicationResourceData(t, nil)
	jsonData, err := prepareApplicationJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.AllowNonPostForm != nil {
		t.Fatalf("expected allow_non_post_form to be omitted for standard, got %v", *jsonData.AllowNonPostForm)
	}
}
