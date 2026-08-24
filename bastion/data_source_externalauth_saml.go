package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExternalAuthSaml() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExternalAuthSamlRead,
		Schema: map[string]*schema.Schema{
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"idp_metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skTimeout: {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			skCertificate: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"claim_customization": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"displayname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						skEmail: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skLanguage: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPassphrase: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"idp_entity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_request_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_request_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_entity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_assertion_consumer_service": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_single_logout_service": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceExternalAuthSamlVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_externalauth_saml not available with api version %s", version)
}

func dataSourceExternalAuthSamlRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthSaml(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get(skAuthenticationName).(string)))
	}
	cfg, err := readExternalAuthSamlOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceExternalAuthSaml(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceExternalAuthSaml(d *schema.ResourceData, jsonData jsonExternalAuthSaml) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_metadata", jsonData.IDPMetadata); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skTimeout, jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_entity_id", jsonData.IDPEntityID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("saml_request_url", jsonData.SamlRequestURL); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("saml_request_method", jsonData.SamlRequestMethod); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_metadata", jsonData.SPMetadata); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_entity_id", jsonData.SPEntityID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_assertion_consumer_service", jsonData.SPAssertionConsumerService); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_single_logout_service", jsonData.SPSingleLogoutService); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.ClaimCustomization != nil {
		claimCustomization := []map[string]interface{}{
			{
				"username":    jsonData.ClaimCustomization.Username,
				"displayname": jsonData.ClaimCustomization.Displayname,
				skEmail:       jsonData.ClaimCustomization.Email,
				skLanguage:    jsonData.ClaimCustomization.Language,
				"group":       jsonData.ClaimCustomization.Group,
			},
		}
		if tfErr := d.Set("claim_customization", claimCustomization); tfErr != nil {
			panic(tfErr)
		}
	}
}
