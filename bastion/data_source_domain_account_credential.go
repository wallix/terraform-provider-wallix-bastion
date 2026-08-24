package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainAccountCredentialRead,
		Schema: map[string]*schema.Schema{
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
			"propagate_credential_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_domain_account_credential not available with api version %s", version)
}

func dataSourceDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomainAccountCredential(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s doesn't exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
	}
	cfg, err := readDomainAccountCredentialOptions(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDomainAccountCredential(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPublicKey, jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
