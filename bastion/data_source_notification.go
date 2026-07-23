package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNotification() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNotificationRead,
		Schema: map[string]*schema.Schema{
			"notification_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"events": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceNotificationVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_notification not available with api version %s", version)
}

func dataSourceNotificationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceNotification(ctx, d.Get("notification_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("notification_name %s doesn't exists", d.Get("notification_name").(string)))
	}
	cfg, err := readNotificationOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceNotification(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceNotification(d *schema.ResourceData, jsonData jsonNotification) {
	if tfErr := d.Set("notification_name", jsonData.NotificationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enabled", jsonData.Enabled); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("destination", jsonData.Destination); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("language", jsonData.Language); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("events", jsonData.Events); tfErr != nil {
		panic(tfErr)
	}
}
