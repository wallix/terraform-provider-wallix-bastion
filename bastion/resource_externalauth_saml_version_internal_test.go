package bastion

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func samlExternalAuthRaw(withClaims bool) map[string]interface{} {
	raw := map[string]interface{}{
		skAuthenticationName: "saml",
		"idp_metadata":       "<md:EntityDescriptor/>",
		skTimeout:            30,
	}
	if withClaims {
		raw["claim_customization"] = []interface{}{map[string]interface{}{"username": "email"}}
	}

	return raw
}

// claim_customization is required on API v3.12 ("'claim_customization' is a required property")
// and refused on v3.8 ("Additional properties are not allowed ('claim_customization' was
// unexpected)"), checked on a Bastion 12.4.1: both are reported at plan.
func TestExternalAuthSamlClaimCustomizationDependsOnTheAPIVersion(t *testing.T) {
	for name, tc := range map[string]struct {
		apiVersion string
		withClaims bool
		wantErr    string
	}{
		"v3.12 with":    {VersionWallixAPI312, true, ""},
		"v3.12 without": {VersionWallixAPI312, false, "claim_customization is required"},
		"v3.8 without":  {VersionWallixAPI38, false, ""},
		"v3.8 with":     {VersionWallixAPI38, true, "claim_customization is not available"},
	} {
		t.Run(name, func(t *testing.T) {
			r := resourceExternalAuthSaml()
			_, err := r.Diff(t.Context(), nil,
				terraform.NewResourceConfigRaw(samlExternalAuthRaw(tc.withClaims)), &Client{bastionAPIVersion: tc.apiVersion})
			switch {
			case tc.wantErr == "" && err != nil:
				t.Errorf("plan error = %v, want none", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Errorf("plan error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

// The body sent to API v3.8 never holds claim_customization, which it refuses.
func TestPrepareExternalAuthSamlJSONClaimCustomization(t *testing.T) {
	r := resourceExternalAuthSaml()
	d := schema.TestResourceDataRaw(t, r.Schema, samlExternalAuthRaw(true))
	for apiVersion, want := range map[string]bool{VersionWallixAPI312: true, VersionWallixAPI38: false} {
		body, err := json.Marshal(prepareExternalAuthSamlJSON(d, apiVersion))
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(body), "claim_customization"); got != want {
			t.Errorf("API %s: claim_customization sent = %v, want %v (%s)", apiVersion, got, want, body)
		}
	}
}
