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
// removal: it is refused at plan. A changed key planned ca_public_key unchanged, so the resources
// using it got the previous key: it is now known after apply.
func TestDomainCAPrivateKeyPlan(t *testing.T) {
	for name, r := range map[string]*schema.Resource{
		"domain": resourceDomain(), "device local domain": resourceDeviceLocalDomain(),
	} {
		for change, tc := range map[string]struct {
			key, description          string
			wantErr, wantNewPublicKey bool
		}{
			"key removed":         {wantErr: true},
			"key changed":         {key: "generate:ECDSA_384", wantNewPublicKey: true},
			"description changed": {key: "ca-pem-key", description: "ca-description"},
		} {
			t.Run(name+" "+change, func(t *testing.T) {
				state := &terraform.InstanceState{ID: "ca-id", Attributes: map[string]string{
					"id": "ca-id", skDeviceID: "ca-device", skDomainName: "ca-domain", skCAPrivateKey: "ca-pem-key",
					skCAPublicKey: "ssh-rsa ca-public-key",
				}}
				config := map[string]interface{}{
					skDeviceID: "ca-device", skDomainName: "ca-domain", skDescription: tc.description,
				}
				if tc.key != "" {
					config[skCAPrivateKey] = tc.key
				}
				diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), nil)
				if (err != nil) != tc.wantErr {
					t.Fatalf("plan error = %v, want an error: %t", err, tc.wantErr)
				}
				if err != nil {
					return
				}
				publicKey, planned := diff.Attributes[skCAPublicKey]
				if got := planned && publicKey.NewComputed; got != tc.wantNewPublicKey {
					t.Errorf("ca_public_key known after apply = %t, want %t", got, tc.wantNewPublicKey)
				}
			})
		}
	}
}
