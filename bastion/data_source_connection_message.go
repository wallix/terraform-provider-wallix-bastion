package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceConnectionMessage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectionMessageRead,
		Schema: map[string]*schema.Schema{
			"message_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"login_en", "login_fr", "login_de", "login_es", "login_ru",
						"motd_en", "motd_fr", "motd_de", "motd_es", "motd_ru",
					},
					false,
				),
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectionMessageVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_connection_message not available with api version %s", version)
}

func dataSourceConnectionMessageRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConnectionMessageVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	messageName := d.Get("message_name").(string)
	cfg, err := readConnectionMessage(ctx, messageName, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.Message == "" {
		return diag.FromErr(fmt.Errorf("message_name %s doesn't exists", messageName))
	}
	cfg.MessageName = messageName
	fillSourceConnectionMessage(d, cfg)
	d.SetId(messageName)

	return nil
}

func fillSourceConnectionMessage(d *schema.ResourceData, jsonData jsonConnectionMessage) {
	if tfErr := d.Set("message_name", jsonData.MessageName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("message", jsonData.Message); tfErr != nil {
		panic(tfErr)
	}
}
