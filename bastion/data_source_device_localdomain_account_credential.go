package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeviceLocalDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceLocalDomainAccountCredentialRead,
		Schema: map[string]*schema.Schema{
			skDeviceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccountID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skType: {
				Type:     schema.TypeString,
				Required: true,
			},
			skPassphrase: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDeviceLocalDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device_localdomain_account_credential "+
		"not available with api version %s", version)
}

func dataSourceDeviceLocalDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
		d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s, device_id %s doesn't exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string), d.Get(skDeviceID).(string)))
	}
	cfg, err := readDeviceLocalDomainAccountCredentialOptions(ctx,
		d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceLocalDomainAccountCredential(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceLocalDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPublicKey, jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
