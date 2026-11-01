package bastion

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	authDomainCATestID     = "authdomain-ca-id"
	authDomainCATestDomain = "authdomain-ca-domain"
	authDomainCATestCA     = "authdomain-ca"
)

func authDomainCAResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{"ldap": resourceAuthDomainLdap(), "ad": resourceAuthDomainAD()}
}

func authDomainCAConfig(extra map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{
		skDomainName: authDomainCATestDomain, skAuthDomainName: "ca.example.com", skDefaultEmailDomain: "example.com",
		skDefaultLanguage: "en", skExternalAuths: []interface{}{"ca-ldap"},
	}
	maps.Copy(config, extra)

	return config
}

// API v3.8 refuses certificate_authority and enable_ca, and the Bastion refuses enable_ca without
// a certificate authority (409): both failed only at apply.
func TestAuthDomainCAPlan(t *testing.T) {
	const (
		unknown     = "74D93920-ED26-11E3-AC10-0800200C9A66" // the SDK's marker of an unknown value
		errRequires = "requires certificate_authority"
		errVersion  = "not available with API version v3.8"
	)
	type cfg = map[string]interface{}
	for name, r := range authDomainCAResources() {
		for desc, tc := range map[string]struct {
			apiVersion string
			config     cfg
			wantErr    string
		}{
			"v3.12 unset":       {VersionWallixAPI312, nil, ""},
			"v3.12 ca enabled":  {VersionWallixAPI312, cfg{skCertificateAuthority: authDomainCATestCA, skEnableCA: true}, ""},
			"v3.12 ca disabled": {VersionWallixAPI312, cfg{skCertificateAuthority: authDomainCATestCA}, ""},
			"v3.12 enabled":     {VersionWallixAPI312, cfg{skEnableCA: true}, errRequires},
			"v3.12 empty ca":    {VersionWallixAPI312, cfg{skCertificateAuthority: "", skEnableCA: true}, errRequires},
			"v3.12 unknown ca":  {VersionWallixAPI312, cfg{skCertificateAuthority: unknown, skEnableCA: true}, ""},
			"v3.8 unset":        {VersionWallixAPI38, nil, ""},
			"v3.8 zero values":  {VersionWallixAPI38, cfg{skCertificateAuthority: "", skEnableCA: false}, ""},
			"v3.8 ca":           {VersionWallixAPI38, cfg{skCertificateAuthority: authDomainCATestCA}, errVersion},
			"v3.8 enabled":      {VersionWallixAPI38, cfg{skEnableCA: true}, errVersion},
			"v3.8 unknown ca":   {VersionWallixAPI38, cfg{skCertificateAuthority: unknown}, errVersion},
		} {
			t.Run(name+" "+desc, func(t *testing.T) {
				config := terraform.NewResourceConfigRaw(authDomainCAConfig(tc.config))
				_, err := r.Diff(t.Context(), &terraform.InstanceState{}, config, &Client{bastionAPIVersion: tc.apiVersion})
				if tc.wantErr == "" && err != nil {
					t.Fatalf("unexpected plan error: %v", err)
				}
				if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
					t.Fatalf("plan error = %v, want %q", err, tc.wantErr)
				}
			})
		}
	}
}

// A state written before these attributes existed, refreshed or not, plans no change for a
// configuration that doesn't set them.
func TestAuthDomainCAUpgradePlansEmpty(t *testing.T) {
	for name, r := range authDomainCAResources() {
		for desc, extra := range map[string]map[string]string{
			"not refreshed": {},
			"refreshed":     {skCertificateAuthority: "", skEnableCA: "false"},
		} {
			t.Run(name+" "+desc, func(t *testing.T) {
				attributes := map[string]string{
					"id": authDomainCATestID, skDomainName: authDomainCATestDomain, skAuthDomainName: "ca.example.com",
					skDefaultEmailDomain: "example.com", skDefaultLanguage: "en", "external_auths.#": "1",
					"external_auths.0": "ca-ldap",
				}
				maps.Copy(attributes, extra)
				diff, err := r.Diff(t.Context(), &terraform.InstanceState{ID: authDomainCATestID, Attributes: attributes},
					terraform.NewResourceConfigRaw(authDomainCAConfig(nil)), &Client{bastionAPIVersion: VersionWallixAPI312})
				if err != nil {
					t.Fatalf("unexpected plan error: %v", err)
				}
				if !diff.Empty() {
					t.Errorf("plan = %v, want no change", diff.Attributes)
				}
			})
		}
	}
}

// An omitted field keeps its value on the Bastion, so v3.12 always sends both; v3.8 refuses them.
func TestPrepareAuthDomainCAJSON(t *testing.T) {
	for desc, tc := range map[string]struct {
		apiVersion string
		config     map[string]interface{}
		want       string
	}{
		"v3.12 set": {
			VersionWallixAPI312,
			map[string]interface{}{skCertificateAuthority: authDomainCATestCA, skEnableCA: true},
			`"certificate_authority":"` + authDomainCATestCA + `","enable_ca":true`,
		},
		"v3.12 unset": {VersionWallixAPI312, nil, `"certificate_authority":"","enable_ca":false`},
		"v3.8 unset":  {VersionWallixAPI38, nil, ""},
	} {
		for name, prepare := range map[string]func(*schema.ResourceData, string) interface{}{
			"ldap": func(d *schema.ResourceData, v string) interface{} { return prepareAuthDomainLdapJSON(d, v) },
			"ad":   func(d *schema.ResourceData, v string) interface{} { return prepareAuthDomainADJSON(d, v) },
		} {
			t.Run(name+" "+desc, func(t *testing.T) {
				r := authDomainCAResources()[name]
				d := schema.TestResourceDataRaw(t, r.Schema, authDomainCAConfig(tc.config))
				body, err := json.Marshal(prepare(d, tc.apiVersion))
				if err != nil {
					t.Fatalf("unexpected marshaling error: %v", err)
				}
				got := string(body)
				if tc.want == "" && (strings.Contains(got, skCertificateAuthority) || strings.Contains(got, skEnableCA)) {
					t.Errorf("body = %s, want neither certificate_authority nor enable_ca", got)
				}
				if tc.want != "" && !strings.Contains(got, tc.want) {
					t.Errorf("body = %s, want %s", got, tc.want)
				}
			})
		}
	}
}

// API v3.8 doesn't return the fields: they read as "" and false.
func TestAuthDomainCAFillWithoutFields(t *testing.T) {
	var cfg jsonAuthDomainLdap
	if err := json.Unmarshal([]byte(`{"id":"ca-id","domain_name":"ca-domain","type":"LDAP"}`), &cfg); err != nil {
		t.Fatalf("unexpected unmarshaling error: %v", err)
	}
	d := resourceAuthDomainLdap().TestResourceData()
	if err := d.Set(skCertificateAuthority, authDomainCATestCA); err != nil {
		t.Fatal(err)
	}
	if err := d.Set(skEnableCA, true); err != nil {
		t.Fatal(err)
	}
	fillAuthDomainLdap(d, cfg)
	if got := d.Get(skCertificateAuthority).(string); got != "" {
		t.Errorf("certificate_authority = %q, want empty", got)
	}
	if got := d.Get(skEnableCA).(bool); got {
		t.Errorf("enable_ca = %t, want false", got)
	}
}
