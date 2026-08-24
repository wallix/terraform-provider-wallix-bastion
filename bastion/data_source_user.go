package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skEmail: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skProfile: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_auths": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"certificate_dn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_change_pwd": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ip_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"preferred_language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_user not available with api version %s", version)
}

func dataSourceUserRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceUserVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	userName := d.Get("user_name").(string)
	ex, err := checkResourceUserExists(ctx, userName, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("user_name %s doesn't exists", userName))
	}
	cfg, err := readUserOptions(ctx, userName, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceUser(d, cfg)
	d.SetId(userName)

	return nil
}

func fillSourceUser(d *schema.ResourceData, jsonData jsonUser) {
	if tfErr := d.Set("user_name", jsonData.UserName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skEmail, jsonData.Email); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skProfile, jsonData.Profile); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user_auths", jsonData.UserAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_dn", jsonData.CertificateCN); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("display_name", jsonData.DisplayName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("expiration_date", jsonData.ExpirationDate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("groups", jsonData.Groups); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_source", jsonData.IPSource); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_disabled", jsonData.IsDisabled); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("preferred_language", jsonData.PreferredLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ssh_public_key", jsonData.SSHPublicKey); tfErr != nil {
		panic(tfErr)
	}
}
