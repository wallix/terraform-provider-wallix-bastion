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
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skProfile: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restrictions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						skAction: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skRules: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skSubprotocol: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			skUsers: {
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
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skProfile, jsonData.Profile); tfErr != nil {
		panic(tfErr)
	}
	restrictions := make([]map[string]interface{}, len(jsonData.Restrictions))
	for i, v := range jsonData.Restrictions {
		restrictions[i] = map[string]interface{}{
			skAction:      v.Action,
			skRules:       v.Rules,
			skSubprotocol: v.SubProtocol,
		}
	}
	if tfErr := d.Set("restrictions", restrictions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsers, jsonData.Users); tfErr != nil {
		panic(tfErr)
	}
}
