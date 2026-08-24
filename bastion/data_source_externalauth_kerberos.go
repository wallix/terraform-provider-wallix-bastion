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
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ker_dom_controller": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kerberos_password": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skDescription: {
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
			skUsePrimaryAuthDomain: {
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
	id, ex, err := searchResourceExternalAuthKerberos(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get(skAuthenticationName).(string)))
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
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ker_dom_controller", jsonData.KerDomController); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Type == "KERBEROS-PASSWORD" {
		if tfErr := d.Set("kerberos_password", true); tfErr != nil {
			panic(tfErr)
		}
	}
}
