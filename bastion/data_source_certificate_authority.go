package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/mod/semver"
)

func dataSourceCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCertificateAuthorityRead,
		Schema: map[string]*schema.Schema{
			"certificate_authority_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ca_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCACertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// certificate_authorities doesn't exist before API v3.12.
func dataSourceCertificateAuthorityVersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_certificate_authority not available with api version %s", version)
}

func dataSourceCertificateAuthorityRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceCertificateAuthority(ctx, d.Get("certificate_authority_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf(
			"certificate_authority_name %s doesn't exists", d.Get("certificate_authority_name").(string)))
	}
	cfg, err := readCertificateAuthorityOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceCertificateAuthority(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceCertificateAuthority(d *schema.ResourceData, jsonData jsonCertificateAuthority) {
	if tfErr := d.Set("certificate_authority_name", jsonData.CertificateAuthorityName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_type", jsonData.CAType); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCACertificate, jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
}
