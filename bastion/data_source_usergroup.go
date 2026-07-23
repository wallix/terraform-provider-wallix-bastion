package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupRead,
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timeframes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restrictions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rules": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subprotocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"users": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceUserGroupVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_usergroup not available with api version %s", version)
}

func dataSourceUserGroupRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceUserGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("group_name %s doesn't exists", d.Get("group_name").(string)))
	}
	cfg, err := readUserGroupOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceUserGroup(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceUserGroup(d *schema.ResourceData, jsonData jsonUserGroup) {
	if tfErr := d.Set("group_name", jsonData.GroupName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("timeframes", jsonData.TimeFrames); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("profile", jsonData.Profile); tfErr != nil {
		panic(tfErr)
	}
	restrictions := make([]map[string]interface{}, len(jsonData.Restrictions))
	for i, v := range jsonData.Restrictions {
		restrictions[i] = map[string]interface{}{
			"action":      v.Action,
			"rules":       v.Rules,
			"subprotocol": v.SubProtocol,
		}
	}
	if tfErr := d.Set("restrictions", restrictions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("users", jsonData.Users); tfErr != nil {
		panic(tfErr)
	}
}
