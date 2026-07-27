package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProfileRead,
		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"gui_features": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"wab_audit": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_audit": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"users": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_groups": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"devices": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_groups": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authorizations": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"profiles": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"wab_settings": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_settings": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"approval": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"credential_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"gui_transmission": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"system_audit": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"users": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_groups": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"devices": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_groups": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authorizations": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"profiles": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"wab_settings": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_settings": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"approval": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"credential_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ip_limitation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_access": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"target_groups_limitation": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_target_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_groups": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"user_groups_limitation": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_groups": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceProfileVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_profile not available with api version %s", version)
}

func dataSourceProfileRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceProfile(ctx, d.Get("profile_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("profile_name %s doesn't exists", d.Get("profile_name").(string)))
	}
	cfg, err := readProfileOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceProfile(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceProfile(d *schema.ResourceData, jsonData jsonProfile) {
	if tfErr := d.Set("profile_name", jsonData.ProfileName); tfErr != nil {
		panic(tfErr)
	}
	guiFeatures := []map[string]interface{}{{
		"wab_audit":           jsonData.GuiFeatures.WabAudit,
		"system_audit":        jsonData.GuiFeatures.SystemAudit,
		"users":               jsonData.GuiFeatures.Users,
		"user_groups":         jsonData.GuiFeatures.UserGroups,
		"devices":             jsonData.GuiFeatures.Devices,
		"target_groups":       jsonData.GuiFeatures.TargetGroups,
		"authorizations":      jsonData.GuiFeatures.Authorizations,
		"profiles":            jsonData.GuiFeatures.Profiles,
		"wab_settings":        jsonData.GuiFeatures.WabSettings,
		"system_settings":     jsonData.GuiFeatures.SystemSettings,
		"backup":              jsonData.GuiFeatures.Backup,
		"approval":            jsonData.GuiFeatures.Approval,
		"credential_recovery": jsonData.GuiFeatures.CredentialRecovery,
	}}
	if tfErr := d.Set("gui_features", guiFeatures); tfErr != nil {
		panic(tfErr)
	}
	guiTransmission := []map[string]interface{}{{
		"system_audit":        jsonData.GuiTransmission.SystemAudit,
		"users":               jsonData.GuiTransmission.Users,
		"user_groups":         jsonData.GuiTransmission.UserGroups,
		"devices":             jsonData.GuiTransmission.Devices,
		"target_groups":       jsonData.GuiTransmission.TargetGroups,
		"authorizations":      jsonData.GuiTransmission.Authorizations,
		"profiles":            jsonData.GuiTransmission.Profiles,
		"wab_settings":        jsonData.GuiTransmission.WabSettings,
		"system_settings":     jsonData.GuiTransmission.SystemSettings,
		"backup":              jsonData.GuiTransmission.Backup,
		"approval":            jsonData.GuiTransmission.Approval,
		"credential_recovery": jsonData.GuiTransmission.CredentialRecovery,
	}}
	if tfErr := d.Set("gui_transmission", guiTransmission); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("dashboards", jsonData.Dashboards); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_limitation", jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("target_access", jsonData.TargetAccess); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.TargetGroupsLimitation.Enabled {
		targetGroupsLimitation := []map[string]interface{}{{
			"default_target_group": *jsonData.TargetGroupsLimitation.DefaultTargetGroup,
			"target_groups":        *jsonData.TargetGroupsLimitation.TargetGroups,
		}}
		if tfErr := d.Set("target_groups_limitation", targetGroupsLimitation); tfErr != nil {
			panic(tfErr)
		}
	} else {
		v := make([]map[string]interface{}, 0)
		if tfErr := d.Set("target_groups_limitation", v); tfErr != nil {
			panic(tfErr)
		}
	}
	if jsonData.UserGroupsLimitation.Enabled {
		userGroupsLimitation := []map[string]interface{}{{
			"user_groups": *jsonData.UserGroupsLimitation.UserGroups,
		}}
		if tfErr := d.Set("user_groups_limitation", userGroupsLimitation); tfErr != nil {
			panic(tfErr)
		}
	} else {
		v := make([]map[string]interface{}, 0)
		if tfErr := d.Set("user_groups_limitation", v); tfErr != nil {
			panic(tfErr)
		}
	}
}
