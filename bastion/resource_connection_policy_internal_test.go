package bastion

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func webappConnectionPolicyRaw() map[string]interface{} {
	return map[string]interface{}{
		"connection_policy_name": "WEBAPP_Custom",
		skProtocol:               skProtoWEBAPP,
		"options":                `{}`,
	}
}

func TestConnectionPolicySchemaAcceptsWEBAPP(t *testing.T) {
	r := resourceConnectionPolicy()
	raw := webappConnectionPolicyRaw()
	raw[skType] = skProtoWEBAPP
	if diags := r.Validate(terraform.NewResourceConfigRaw(raw)); diags.HasError() {
		t.Errorf("protocol and type WEBAPP must be valid: %v", diags)
	}

	// Guard against a validation that accepts anything.
	raw[skProtocol] = "FOO"
	if diags := r.Validate(terraform.NewResourceConfigRaw(raw)); !diags.HasError() {
		t.Error("protocol FOO must be rejected")
	}
}

func TestPrepareConnectionPolicyJSONSendsWEBAPP(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceConnectionPolicy().Schema, webappConnectionPolicyRaw())

	jsonData, err := prepareConnectionPolicyJSON(d, true, VersionWallixAPI312)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Protocol != skProtoWEBAPP || jsonData.Type != skProtoWEBAPP {
		t.Errorf("protocol, type = %q, %q, want WEBAPP, WEBAPP", jsonData.Protocol, jsonData.Type)
	}
}

func TestPrepareConnectionPolicyJSONRejectsWEBAPPBeforeAPI312(t *testing.T) {
	for name, raw := range map[string]map[string]interface{}{
		"protocol": webappConnectionPolicyRaw(),
		"type": {
			"connection_policy_name": "WEBAPP_Custom",
			skProtocol:               skProtoRDP,
			skType:                   skProtoWEBAPP,
			"options":                `{}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceConnectionPolicy().Schema, raw)
			_, err := prepareConnectionPolicyJSON(d, true, VersionWallixAPI38)
			if err == nil || !strings.Contains(err.Error(), skProtoWEBAPP) {
				t.Errorf("error = %v, want the WEBAPP version error with api version v3.8", err)
			}
		})
	}
}

// The WEBAPP version guard must not reject the other protocols before API v3.12.
func TestPrepareConnectionPolicyJSONCreatesSSHBeforeAPI312(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceConnectionPolicy().Schema, map[string]interface{}{
		"connection_policy_name": "SSH_Custom",
		skProtocol:               skProtoSSH,
		"options":                `{}`,
	})

	jsonData, err := prepareConnectionPolicyJSON(d, true, VersionWallixAPI38)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// type is only sent from API v3.12.
	if jsonData.Protocol != skProtoSSH || jsonData.Type != "" {
		t.Errorf("protocol, type = %q, %q, want SSH and no type", jsonData.Protocol, jsonData.Type)
	}
}

func TestPrepareConnectionPolicyJSONUpdateIgnoresWEBAPPVersionCheck(t *testing.T) {
	// protocol and type are only sent on creation: an update never sends them, so it must not
	// be rejected (the resource already exists on the Bastion).
	d := schema.TestResourceDataRaw(t, resourceConnectionPolicy().Schema, webappConnectionPolicyRaw())

	jsonData, err := prepareConnectionPolicyJSON(d, false, VersionWallixAPI38)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonData.Protocol != "" || jsonData.Type != "" {
		t.Errorf("protocol, type = %q, %q, want them omitted on update", jsonData.Protocol, jsonData.Type)
	}
}
