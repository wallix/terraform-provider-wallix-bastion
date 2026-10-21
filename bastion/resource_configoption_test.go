package bastion_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// Options as returned by GET /configoptions/wabengine.
const testConfigoptionGetOptions = `[
  {"name": "one_time_password_ttl", "value": 30.0},
  {"name": "fqdn", "value": "bastion.example.com"},
  {"name": "dict_option", "value": {"a": 1, "b": 2}},
  {"name": "am", "options": [
    {"name": "session_timeout", "value": 1200},
    {"name": "other", "value": true}
  ]}
]`

func testConfigoptionValuesJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func TestConfigoptionValues(t *testing.T) {
	var options []interface{}
	if err := json.Unmarshal([]byte(testConfigoptionGetOptions), &options); err != nil {
		t.Fatal(err)
	}
	values := bastion.ConfigOptionsToValues(options)

	tests := []struct {
		name    string
		managed string
		want    string
	}{
		{
			name:    "all options",
			managed: "",
			want: `{"am":{"other":true,"session_timeout":1200},"dict_option":{"a":1,"b":2},` +
				`"fqdn":"bastion.example.com","one_time_password_ttl":30}`,
		},
		{
			name:    "option in section",
			managed: `{"am":{"session_timeout":2000}}`,
			want:    `{"am":{"session_timeout":1200}}`,
		},
		{
			name:    "option value is kept whole",
			managed: `{"dict_option":{"a":3}}`,
			want:    `{"dict_option":{"a":1,"b":2}}`,
		},
		{
			name:    "missing option is dropped",
			managed: `{"fqdn":"x","unknown":1}`,
			want:    `{"fqdn":"bastion.example.com"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := values
			if tt.managed != "" {
				var managed map[string]interface{}
				if err := json.Unmarshal([]byte(tt.managed), &managed); err != nil {
					t.Fatal(err)
				}
				got = bastion.FilterConfigValues(values, managed)
			}
			if g := testConfigoptionValuesJSON(t, got); g != tt.want {
				t.Errorf("got %s, want %s", g, tt.want)
			}
		})
	}
}

func TestConfigoptionPrepareValues(t *testing.T) {
	var options []interface{}
	if err := json.Unmarshal([]byte(`[
  {"name": "one_time_password_ttl", "type": "float", "value": 30.0},
  {"name": "ratio", "type": "float", "value": 0.5},
  {"name": "count", "type": "integer", "value": 3},
  {"name": "rdp_color_bpp", "type": "option", "value": "16"},
  {"name": "older_than", "type": "string", "value": "5"},
  {"name": "am", "options": [
    {"name": "delay", "type": "float", "value": 1.0},
    {"name": "session_timeout", "type": "integer", "value": 1200}
  ]}
]`), &options); err != nil {
		t.Fatal(err)
	}
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(
		`{"one_time_password_ttl":45,"ratio":0.25,"count":4,"rdp_color_bpp":32,"older_than":2.5,`+
			`"am":{"delay":2,"session_timeout":1500}}`,
	), &values); err != nil {
		t.Fatal(err)
	}
	got, err := bastion.MarshalConfigValuesForPut(values, options)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"am":{"delay":2.0,"session_timeout":1500},"count":4,"older_than":"2.5",` +
		`"one_time_password_ttl":45.0,"ratio":0.25,"rdp_color_bpp":"32"}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}

	if _, err := bastion.MarshalConfigValuesForPut(map[string]interface{}{"am": "x"}, options); err == nil ||
		err.Error() != "am is a section, it must be set with an object of options" {
		t.Errorf("got error %v, want section error", err)
	}
	if _, err := bastion.MarshalConfigValuesForPut(
		map[string]interface{}{"am": map[string]interface{}{"delay": map[string]interface{}{"a": 1.0}}}, options,
	); err == nil || err.Error() != "am.delay is an option, it can't be set with an object" {
		t.Errorf("got error %v, want option error", err)
	}
}

func TestConfigoptionAddPath(t *testing.T) {
	managed := make(map[string]interface{})
	bastion.AddConfigValuesPath(managed, []string{"fqdn"})
	bastion.AddConfigValuesPath(managed, []string{"am", "session_timeout"})
	bastion.AddConfigValuesPath(managed, []string{"am", "other"})
	want := `{"am":{"other":null,"session_timeout":null},"fqdn":null}`
	if g := testConfigoptionValuesJSON(t, managed); g != want {
		t.Errorf("got %s, want %s", g, want)
	}
}

func TestConfigoptionSuppressEquivalentJSON(t *testing.T) {
	tests := []struct {
		old, new string
		want     bool
	}{
		{`{"a":30.0,"b":{"c":true}}`, "{\n  \"b\": {\"c\": true},\n  \"a\": 30\n}", true},
		{`{"a":30}`, `{"a":31}`, false},
		{`{"a":"16","b":{"c":"2.5"}}`, `{"a":16,"b":{"c":2.5}}`, true},
		{`{"a":"16"}`, `{"a":17}`, false},
		{`{"a":"16"}`, `{"a":16,"b":1}`, false},
		{``, `{"a":30}`, false},
	}
	for _, tt := range tests {
		if got := bastion.SuppressEquivalentJSON(tt.old, tt.new); got != tt.want {
			t.Errorf("SuppressEquivalentJSON(%q, %q) = %v, want %v", tt.old, tt.new, got, tt.want)
		}
	}
}

// PUT on configoptions doesn't exist before API v3.12 and requires the REST API option
// preview_resources to be enabled on the Bastion.
func TestAccResourceConfigoption_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resourceName := "wallix-bastion_configoption.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceConfigoptionConfig(45, 1500),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(resourceName, "id",
							"data.wallix-bastion_configoption.wabengine", "id"),
						resource.TestCheckResourceAttr(resourceName, "config_name", "wabengine"),
						resource.TestCheckResourceAttr(resourceName, "name", "Global"),
						resource.TestCheckResourceAttr(resourceName, "values",
							`{"am":{"session_timeout":1500},"one_time_password_ttl":45}`),
					),
				},
				{
					Config: testAccResourceConfigoptionConfig(30, 1200),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "values",
							`{"am":{"session_timeout":1200},"one_time_password_ttl":30}`),
					),
				},
				{
					// Same configuration set by id instead of name: no change.
					Config:   testAccResourceConfigoptionConfigByID(30, 1200),
					PlanOnly: true,
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateId:     "wabengine:one_time_password_ttl,am.session_timeout",
					ImportStateVerify: true,
				},
			},
			PreventPostDestroyRefresh: true, // Configuration always exists on the Bastion; nothing to destroy.
		})
	}
}

func testAccResourceConfigoptionConfig(ttl, sessionTimeout int) string {
	return fmt.Sprintf(`
data "wallix-bastion_configoption" "wabengine" {
  config_id    = "wabengine"
  options_list = ["one_time_password_ttl"]
}

resource "wallix-bastion_configoption" "test" {
  config_id = "wabengine"
  values = jsonencode({
    one_time_password_ttl = %d
    am = {
      session_timeout = %d
    }
  })
}
`, ttl, sessionTimeout)
}

func testAccResourceConfigoptionConfigByID(ttl, sessionTimeout int) string {
	return fmt.Sprintf(`
data "wallix-bastion_configoption" "wabengine" {
  config_id    = "wabengine"
  options_list = ["one_time_password_ttl"]
}

resource "wallix-bastion_configoption" "test" {
  config_id = data.wallix-bastion_configoption.wabengine.id
  values = jsonencode({
    one_time_password_ttl = %d
    am = {
      session_timeout = %d
    }
  })
}
`, ttl, sessionTimeout)
}
