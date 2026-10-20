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
		// The API returns them in clear: wabengine backup_key, wabrestapi data_encryption_key,
		// sessionlogpolicy archive_key (checked on a Bastion 12.4), and the passwords of the
		// options (proxy.password, vnc_over_ssh.ssh_password, rdp.krb_armoring_fallback_password).
		"resource configoption.values":          resourceConfigoption().Schema[skValues],
		"data source configoption.values":       dataSourceConfigoption().Schema[skValues],
		"data source configoption.options":      dataSourceConfigoption().Schema["options"],
		"resource connection_policy.options":    resourceConnectionPolicy().Schema["options"],
		"data source connection_policy.options": dataSourceConnectionPolicy().Schema["options"],
	} {
		if attribute == nil || !attribute.Sensitive {
			t.Errorf("%s must be sensitive", name)
		}
	}
}
