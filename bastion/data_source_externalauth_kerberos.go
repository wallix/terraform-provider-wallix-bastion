package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExternalAuthKerberos() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExternalAuthKerberosRead,
		Schema: map[string]*schema.Schema{
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ker_dom_controller": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kerberos_password": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keytab": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"login_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"use_primary_auth_domain": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceExternalAuthKerberosVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_externalauth_kerberos not available with api version %s", version)
}

func dataSourceExternalAuthKerberosRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthKerberos(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get("authentication_name").(string)))
	}
	cfg, err := readExternalAuthKerberosOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceExternalAuthKerberos(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceExternalAuthKerberos(d *schema.ResourceData, jsonData jsonExternalAuthKerberos) {
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ker_dom_controller", jsonData.KerDomController); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Type == "KERBEROS-PASSWORD" {
		if tfErr := d.Set("kerberos_password", true); tfErr != nil {
			panic(tfErr)
		}
	}
}
