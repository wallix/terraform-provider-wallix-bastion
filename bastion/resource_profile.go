package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonProfile struct {
	TargetAccess bool      `json:"target_access"`
	ID           string    `json:"id,omitempty"`
	ProfileName  string    `json:"profile_name,omitempty"`
	Description  string    `json:"description"`
	IPLimitation string    `json:"ip_limitation"`
	Dashboards   *[]string `json:"dashboards,omitempty"`
	GuiFeatures  struct {
		WabAudit           *string `json:"wab_audit"`
		SystemAudit        *string `json:"system_audit"`
		Users              *string `json:"users"`
		UserGroups         *string `json:"user_groups"`
		Devices            *string `json:"devices"`
		TargetGroups       *string `json:"target_groups"`
		Authorizations     *string `json:"authorizations"`
		Profiles           *string `json:"profiles"`
		WabSettings        *string `json:"wab_settings"`
		SystemSettings     *string `json:"system_settings"`
		Backup             *string `json:"backup"`
		Approval           *string `json:"approval"`
		CredentialRecovery *string `json:"credential_recovery"`
	} `json:"gui_features"`
	GuiTransmission struct {
		SystemAudit        *string `json:"system_audit"`
		Users              *string `json:"users"`
		UserGroups         *string `json:"user_groups"`
		Devices            *string `json:"devices"`
		TargetGroups       *string `json:"target_groups"`
		Authorizations     *string `json:"authorizations"`
		Profiles           *string `json:"profiles"`
		WabSettings        *string `json:"wab_settings"`
		SystemSettings     *string `json:"system_settings"`
		Backup             *string `json:"backup"`
		Approval           *string `json:"approval"`
		CredentialRecovery *string `json:"credential_recovery"`
	} `json:"gui_transmission"`
	TargetGroupsLimitation struct {
		Enabled            bool         `json:"enabled"`
		DefaultTargetGroup *interface{} `json:"default_target_group,omitempty"`
		TargetGroups       *[]string    `json:"target_groups,omitempty"`
	} `json:"target_groups_limitation"`
	UserGroupsLimitation struct {
		Enabled    bool      `json:"enabled"`
		UserGroups *[]string `json:"user_groups,omitempty"`
	} `json:"user_groups_limitation"`
}

func resourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProfileCreate,
		ReadContext:   resourceProfileRead,
		UpdateContext: resourceProfileUpdate,
		DeleteContext: resourceProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceProfileImport,
		},
		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gui_features": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"wab_audit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view"}, false),
						},
						"system_audit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view"}, false),
						},
						"users": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"user_groups": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"devices": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"target_groups": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"authorizations": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"profiles": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"modify"}, false),
						},
						"wab_settings": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"system_settings": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"modify"}, false),
						},
						"backup": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"execute"}, false),
						},
						"approval": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"credential_recovery": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"execute"}, false),
						},
					},
				},
			},
			"gui_transmission": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"system_audit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view"}, false),
						},
						"users": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"user_groups": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"devices": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"target_groups": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"authorizations": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"profiles": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"modify"}, false),
						},
						"wab_settings": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"system_settings": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"modify"}, false),
						},
						"backup": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"execute"}, false),
						},
						"approval": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"view", "modify"}, false),
						},
						"credential_recovery": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"execute"}, false),
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dashboards": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ip_limitation": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"target_groups_limitation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_target_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_groups": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"user_groups_limitation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_groups": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceProfileVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_profile not available with api version %s", version)
}

func resourceProfileCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceProfile(ctx, d.Get("profile_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("profile_name %s already exists", d.Get("profile_name").(string)))
	}
	err = addProfile(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceProfile(ctx, d.Get("profile_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("profile_name %s not found after POST", d.Get("profile_name").(string)))
	}
	d.SetId(id)

	return resourceProfileRead(ctx, d, m)
}

func resourceProfileRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readProfileOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillProfile(d, cfg)
	}

	return nil
}

func resourceProfileUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateProfile(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceProfileRead(ctx, d, m)
}

func resourceProfileDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteProfile(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProfileImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceProfileVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceProfile(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find profile_name with id %s (id must be <profile_name>)", d.Id())
	}
	cfg, err := readProfileOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillProfile(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("profile_name", d.Id()); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceProfile(
	ctx context.Context, profileName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/profiles/?q=profile_name="+profileName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonProfile
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addProfile(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareProfileJSON(d, true)
	body, code, err := c.newRequest(ctx, "/profiles/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateProfile(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareProfileJSON(d, false)
	body, code, err := c.newRequest(ctx, "/profiles/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteProfile(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/profiles/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareProfileJSON( //nolint: gocognit,gocyclo
	d *schema.ResourceData, newResource bool,
) jsonProfile {
	jsonData := jsonProfile{
		Description:  d.Get("description").(string),
		IPLimitation: d.Get("ip_limitation").(string),
		TargetAccess: d.Get("target_access").(bool),
	}

	if newResource {
		jsonData.ProfileName = d.Get("profile_name").(string)
	}

	for _, v := range d.Get("gui_features").([]interface{}) {
		if v == nil {
			continue
		}
		m := v.(map[string]interface{})
		if v2 := m["wab_audit"].(string); v2 != "" {
			jsonData.GuiFeatures.WabAudit = &v2
		}
		if v2 := m["system_audit"].(string); v2 != "" {
			jsonData.GuiFeatures.SystemAudit = &v2
		}
		if v2 := m["users"].(string); v2 != "" {
			jsonData.GuiFeatures.Users = &v2
		}
		if v2 := m["user_groups"].(string); v2 != "" {
			jsonData.GuiFeatures.UserGroups = &v2
		}
		if v2 := m["devices"].(string); v2 != "" {
			jsonData.GuiFeatures.Devices = &v2
		}
		if v2 := m["target_groups"].(string); v2 != "" {
			jsonData.GuiFeatures.TargetGroups = &v2
		}
		if v2 := m["authorizations"].(string); v2 != "" {
			jsonData.GuiFeatures.Authorizations = &v2
		}
		if v2 := m["profiles"].(string); v2 != "" {
			jsonData.GuiFeatures.Profiles = &v2
		}
		if v2 := m["wab_settings"].(string); v2 != "" {
			jsonData.GuiFeatures.WabSettings = &v2
		}
		if v2 := m["system_settings"].(string); v2 != "" {
			jsonData.GuiFeatures.SystemSettings = &v2
		}
		if v2 := m["backup"].(string); v2 != "" {
			jsonData.GuiFeatures.Backup = &v2
		}
		if v2 := m["approval"].(string); v2 != "" {
			jsonData.GuiFeatures.Approval = &v2
		}
		if v2 := m["credential_recovery"].(string); v2 != "" {
			jsonData.GuiFeatures.CredentialRecovery = &v2
		}
	}

	for _, v := range d.Get("gui_transmission").([]interface{}) {
		if v == nil {
			continue
		}
		m := v.(map[string]interface{})
		if v2 := m["system_audit"].(string); v2 != "" {
			jsonData.GuiTransmission.SystemAudit = &v2
		}
		if v2 := m["users"].(string); v2 != "" {
			jsonData.GuiTransmission.Users = &v2
		}
		if v2 := m["user_groups"].(string); v2 != "" {
			jsonData.GuiTransmission.UserGroups = &v2
		}
		if v2 := m["devices"].(string); v2 != "" {
			jsonData.GuiTransmission.Devices = &v2
		}
		if v2 := m["target_groups"].(string); v2 != "" {
			jsonData.GuiTransmission.TargetGroups = &v2
		}
		if v2 := m["authorizations"].(string); v2 != "" {
			jsonData.GuiTransmission.Authorizations = &v2
		}
		if v2 := m["profiles"].(string); v2 != "" {
			jsonData.GuiTransmission.Profiles = &v2
		}
		if v2 := m["wab_settings"].(string); v2 != "" {
			jsonData.GuiTransmission.WabSettings = &v2
		}
		if v2 := m["system_settings"].(string); v2 != "" {
			jsonData.GuiTransmission.SystemSettings = &v2
		}
		if v2 := m["backup"].(string); v2 != "" {
			jsonData.GuiTransmission.Backup = &v2
		}
		if v2 := m["approval"].(string); v2 != "" {
			jsonData.GuiTransmission.Approval = &v2
		}
		if v2 := m["credential_recovery"].(string); v2 != "" {
			jsonData.GuiTransmission.CredentialRecovery = &v2
		}
	}

	listDashboards := d.Get("dashboards").(*schema.Set).List()
	if len(listDashboards) > 0 {
		dashboards := make([]string, len(listDashboards))
		for i, v := range listDashboards {
			dashboards[i] = v.(string)
		}
		jsonData.Dashboards = &dashboards
	}

	for _, v := range d.Get("target_groups_limitation").([]interface{}) {
		m := v.(map[string]interface{})
		jsonData.TargetGroupsLimitation.Enabled = true
		listTargetGroups := m["target_groups"].(*schema.Set).List()
		targetGroups := make([]string, len(listTargetGroups))
		for i, v2 := range listTargetGroups {
			targetGroups[i] = v2.(string)
		}
		jsonData.TargetGroupsLimitation.TargetGroups = &targetGroups
		var defaultTargetGroup interface{}
		if v2 := m["default_target_group"].(string); v2 != "" {
			defaultTargetGroup = v2
		}
		jsonData.TargetGroupsLimitation.DefaultTargetGroup = &defaultTargetGroup
	}

	for _, v := range d.Get("user_groups_limitation").([]interface{}) {
		m := v.(map[string]interface{})
		jsonData.UserGroupsLimitation.Enabled = true
		listUserGroups := m["user_groups"].(*schema.Set).List()
		userGroups := make([]string, len(listUserGroups))
		for i, v2 := range listUserGroups {
			userGroups[i] = v2.(string)
		}
		jsonData.UserGroupsLimitation.UserGroups = &userGroups
	}

	return jsonData
}

func readProfileOptions(
	ctx context.Context, profileID string, m interface{},
) (
	jsonProfile, error,
) {
	c := m.(*Client)
	var result jsonProfile
	body, code, err := c.newRequest(ctx, "/profiles/"+profileID, http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}

func fillProfile(d *schema.ResourceData, jsonData jsonProfile) {
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
	} else if _, ok := d.GetOk("target_groups_limitation"); ok {
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
	} else if _, ok := d.GetOk("user_groups_limitation"); ok {
		v := make([]map[string]interface{}, 0)
		if tfErr := d.Set("user_groups_limitation", v); tfErr != nil {
			panic(tfErr)
		}
	}
}
