package bastion

import (
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// A new generate: value of ca_private_key was only sent when there was no key before, so the
// Bastion kept its key while the state recorded the new value. Sending an unchanged generate:
// value would regenerate the key on every update.
func TestPrepareDomainJSONSendsChangedGeneratedCAKey(t *testing.T) {
	for name, tc := range map[string]struct {
		resource *schema.Resource
		raw      map[string]interface{}
		caKey    func(d *schema.ResourceData) string
	}{
		"domain": {
			resource: resourceDomain(),
			raw:      map[string]interface{}{skDomainName: "ca-domain"},
			caKey:    func(d *schema.ResourceData) string { return prepareDomainJSON(d, false).CAPrivateKey },
		},
		"device local domain": {
			resource: resourceDeviceLocalDomain(),
			raw:      map[string]interface{}{skDeviceID: "ca-device", skDomainName: "ca-domain"},
			caKey: func(d *schema.ResourceData) string {
				return prepareDeviceLocalDomainJSON(d, false).CAPrivateKey
			},
		},
	} {
		for _, step := range []struct{ oldKey, newKey, want string }{
			{"generate:RSA_2048", "generate:ED25519", "generate:ED25519"},
			{"ca-pem-key", "generate:ED25519", "generate:ED25519"},
			{"generate:RSA_2048", "generate:RSA_2048", ""},
		} {
			t.Run(name+" "+step.oldKey+" to "+step.newKey, func(t *testing.T) {
				attributes := map[string]string{"id": "ca-id", skCAPrivateKey: step.oldKey}
				for k, v := range tc.raw {
					attributes[k] = v.(string)
				}
				state := &terraform.InstanceState{ID: "ca-id", Attributes: attributes}
				raw := maps.Clone(tc.raw)
				raw[skCAPrivateKey] = step.newKey
				raw[skDescription] = "ca-update"

				sm := schema.InternalMap(tc.resource.Schema)
				diff, err := sm.Diff(t.Context(), state, terraform.NewResourceConfigRaw(raw), nil, nil, true)
				if err != nil {
					t.Fatalf("unexpected error computing the diff: %v", err)
				}
				d, err := sm.Data(state, diff)
				if err != nil {
					t.Fatalf("unexpected error building the resource data: %v", err)
				}
				if got := tc.caKey(d); got != step.want {
					t.Errorf("ca_private_key = %q, want %q", got, step.want)
				}
			})
		}
	}
}

// Removing ca_private_key left the CA key active on the Bastion while the state recorded the
// removal: it is refused at plan, and replacing the key is still planned.
func TestDomainCAPrivateKeyRemovalRefused(t *testing.T) {
	for name, r := range map[string]*schema.Resource{
		"domain": resourceDomain(), "device local domain": resourceDeviceLocalDomain(),
	} {
		for newKey, wantErr := range map[string]bool{"": true, "generate:ECDSA_384": false} {
			t.Run(name+" to "+newKey, func(t *testing.T) {
				state := &terraform.InstanceState{ID: "ca-id", Attributes: map[string]string{
					"id": "ca-id", skDeviceID: "ca-device", skDomainName: "ca-domain", skCAPrivateKey: "ca-pem-key",
				}}
				config := map[string]interface{}{skDeviceID: "ca-device", skDomainName: "ca-domain"}
				if newKey != "" {
					config[skCAPrivateKey] = newKey
				}
				if _, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), nil); (err != nil) != wantErr {
					t.Errorf("plan error = %v, want an error: %t", err, wantErr)
				}
			})
		}
	}
}
