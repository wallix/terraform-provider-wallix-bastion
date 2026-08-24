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
						skSystemAudit: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skUsers: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skUserGroups: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDevices: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skTargetGroups: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skAuthorizations: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skProfiles: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skWabSettings: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skSystemSettings: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skBackup: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skApproval: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skCredentialRecovery: {
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
						skSystemAudit: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skUsers: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skUserGroups: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDevices: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skTargetGroups: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skAuthorizations: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skProfiles: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skWabSettings: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skSystemSettings: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skBackup: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skApproval: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skCredentialRecovery: {
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
			"dashboards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skIPLimitation: {
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
						skTargetGroups: {
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
						skUserGroups: {
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
		"wab_audit":          jsonData.GuiFeatures.WabAudit,
		skSystemAudit:        jsonData.GuiFeatures.SystemAudit,
		skUsers:              jsonData.GuiFeatures.Users,
		skUserGroups:         jsonData.GuiFeatures.UserGroups,
		skDevices:            jsonData.GuiFeatures.Devices,
		skTargetGroups:       jsonData.GuiFeatures.TargetGroups,
		skAuthorizations:     jsonData.GuiFeatures.Authorizations,
		skProfiles:           jsonData.GuiFeatures.Profiles,
		skWabSettings:        jsonData.GuiFeatures.WabSettings,
		skSystemSettings:     jsonData.GuiFeatures.SystemSettings,
		skBackup:             jsonData.GuiFeatures.Backup,
		skApproval:           jsonData.GuiFeatures.Approval,
		skCredentialRecovery: jsonData.GuiFeatures.CredentialRecovery,
	}}
	if tfErr := d.Set("gui_features", guiFeatures); tfErr != nil {
		panic(tfErr)
	}
	guiTransmission := []map[string]interface{}{{
		skSystemAudit:        jsonData.GuiTransmission.SystemAudit,
		skUsers:              jsonData.GuiTransmission.Users,
		skUserGroups:         jsonData.GuiTransmission.UserGroups,
		skDevices:            jsonData.GuiTransmission.Devices,
		skTargetGroups:       jsonData.GuiTransmission.TargetGroups,
		skAuthorizations:     jsonData.GuiTransmission.Authorizations,
		skProfiles:           jsonData.GuiTransmission.Profiles,
		skWabSettings:        jsonData.GuiTransmission.WabSettings,
		skSystemSettings:     jsonData.GuiTransmission.SystemSettings,
		skBackup:             jsonData.GuiTransmission.Backup,
		skApproval:           jsonData.GuiTransmission.Approval,
		skCredentialRecovery: jsonData.GuiTransmission.CredentialRecovery,
	}}
	if tfErr := d.Set("gui_transmission", guiTransmission); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("dashboards", jsonData.Dashboards); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skIPLimitation, jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("target_access", jsonData.TargetAccess); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.TargetGroupsLimitation.Enabled {
		targetGroupsLimitation := []map[string]interface{}{{
			"default_target_group": *jsonData.TargetGroupsLimitation.DefaultTargetGroup,
			skTargetGroups:         *jsonData.TargetGroupsLimitation.TargetGroups,
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
			skUserGroups: *jsonData.UserGroupsLimitation.UserGroups,
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
