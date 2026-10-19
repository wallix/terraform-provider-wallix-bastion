package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceEncryption exposes the singleton encryption/vault status of the Bastion.
//
// The encryption resource does not have a natural per-instance name: the API stores a
// single encryption configuration reachable at the static ID "encryption" (see
// resource_encryption.go, resourceEncryptionImport). Additionally, the API never returns
// the passphrase back (it's write-only), and there is no read...Options function on the
// resource that decodes a jsonEncryption from the API: the resource's own Read only calls
// verifyEncryption to check whether encryption is configured/ready. This data source mirrors
// that: it takes no lookup argument and reports whether encryption is enabled/ready.
func dataSourceEncryption() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEncryptionRead,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"seal_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEncryptionVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_encryption not available with api version %s", version)
}

func dataSourceEncryptionRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceEncryptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	state, err := readEncryptionState(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceEncryption(d, state)
	d.SetId("encryption")

	return nil
}

// enabled is whether the encryption is set up, sealed or not: seal_state tells whether the
// Bastion waits for its passphrase.
func fillSourceEncryption(d *schema.ResourceData, state jsonEncryptionState) {
	if tfErr := d.Set("enabled", state.SealState != encryptionNeedSetup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("seal_state", state.SealState); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("encryption_mode", state.EncryptionMode); tfErr != nil {
		panic(tfErr)
	}
}
