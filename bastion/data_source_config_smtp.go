package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// config_smtp is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "smtpConfig", the same one
// used by resourceConfigSMTP). There is no natural per-instance name to look
// up, so this data source has no Required lookup field: it always reads the
// single existing SMTP configuration.
func dataSourceConfigSMTP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigSMTPRead,
		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"postmaster_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sender_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sender_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigSMTPVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_config_smtp not available with api version %s", version)
}

func dataSourceConfigSMTPRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigSMTPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigSMTPOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceConfigSMTP(d, cfg)
	d.SetId("smtpConfig")

	return nil
}

func fillSourceConfigSMTP(d *schema.ResourceData, jsonData jsonConfigSMTP) {
	if tfErr := d.Set("protocol", jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authentication_method", jsonData.AuthenticationMethod); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("server", jsonData.Server); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("postmaster_email", jsonData.PostmasterEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sender_name", jsonData.SenderName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sender_email", jsonData.SenderEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_hash", jsonData.CertificateHash); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user", jsonData.User); tfErr != nil {
		panic(tfErr)
	}
}
