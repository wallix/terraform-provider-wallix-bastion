package bastion

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// These secrets were shown in plans (and CI logs) because they weren't marked sensitive.
func TestSecretsAreSensitive(t *testing.T) {
	for name, attribute := range map[string]*schema.Schema{
		"resource encryption.current_passphrase":     resourceEncryption().Schema["current_passphrase"],
		"resource encryption.new_passphrase":         resourceEncryption().Schema["new_passphrase"],
		"resource config_x509.server_private_key":    resourceConfigX509().Schema["server_private_key"],
		"data source config_x509.server_private_key": dataSourceConfigX509().Schema["server_private_key"],
	} {
		if attribute == nil || !attribute.Sensitive {
			t.Errorf("%s must be sensitive", name)
		}
	}
}
