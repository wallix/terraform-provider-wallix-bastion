package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonUserGroup struct {
	Users        *[]string         `json:"users,omitempty"`
	ID           string            `json:"id,omitempty"`
	Description  string            `json:"description"`
	GroupName    string            `json:"group_name"`
	Profile      string            `json:"profile"`
	TimeFrames   []string          `json:"timeframes"`
	Restrictions []jsonRestriction `json:"restrictions"`
}

func resourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timeframes": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"restrictions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"kill", "notify"}, false),
						},
						"rules": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subprotocol": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"SSH_SHELL_SESSION", "SSH_REMOTE_COMMAND", "SSH_SCP_UP", "SSH_SCP_DOWN",
								"SFTP_SESSION", "RLOGIN", "TELNET", "RDP"},
								false),
						},
					},
				},
			},
			"users": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}
func resourceUserGroupVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_usergroup not available with api version %s", version)
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceUserGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("group_name %s already exists", d.Get("group_name").(string)))
	}
	err = addUserGroup(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceUserGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("group_name %s not found after POST", d.Get("group_name").(string)))
	}
	d.SetId(id)

	return resourceUserGroupRead(ctx, d, m)
}
func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readUserGroupOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillUserGroup(d, cfg)
	}

	return nil
}
func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateUserGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceUserGroupRead(ctx, d, m)
}
func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteUserGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceUserGroupImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceUserGroup(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find group_name with id %s (id must be <group_name>", d.Id())
	}
	cfg, err := readUserGroupOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillUserGroup(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceUserGroup(ctx context.Context, groupName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/usergroups/?fields=group_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonUserGroup
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.GroupName == groupName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addUserGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareUserGroupJSON(d)
	body, code, err := c.newRequest(ctx, "/usergroups/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateUserGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareUserGroupJSON(d)
	body, code, err := c.newRequest(ctx, "/usergroups/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteUserGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/usergroups/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareUserGroupJSON(d *schema.ResourceData) jsonUserGroup {
	jsonData := jsonUserGroup{
		Description: d.Get("description").(string),
		GroupName:   d.Get("group_name").(string),
		Profile:     d.Get("profile").(string),
	}
	if d.HasChanges("users") {
		users := make([]string, 0)
		for _, v := range d.Get("users").(*schema.Set).List() {
			users = append(users, v.(string))
		}
		jsonData.Users = &users
	}
	for _, v := range d.Get("timeframes").(*schema.Set).List() {
		jsonData.TimeFrames = append(jsonData.TimeFrames, v.(string))
	}
	if len(d.Get("restrictions").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("restrictions").(*schema.Set).List() {
			r := v.(map[string]interface{})
			jsonData.Restrictions = append(jsonData.Restrictions, jsonRestriction{
				Action:      r["action"].(string),
				Rules:       r["rules"].(string),
				SubProtocol: r["subprotocol"].(string),
			})
		}
	} else {
		jsonData.Restrictions = make([]jsonRestriction, 0)
	}

	return jsonData
}

func readUserGroupOptions(
	ctx context.Context, groupID string, m interface{}) (jsonUserGroup, error) {
	c := m.(*Client)
	var result jsonUserGroup
	body, code, err := c.newRequest(ctx, "/usergroups/"+groupID, http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillUserGroup(d *schema.ResourceData, jsonData jsonUserGroup) {
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
	restrictions := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Restrictions {
		restrictions = append(restrictions, map[string]interface{}{
			"action":      v.Action,
			"rules":       v.Rules,
			"subprotocol": v.SubProtocol,
		})
	}
	if tfErr := d.Set("restrictions", restrictions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("users", jsonData.Users); tfErr != nil {
		panic(tfErr)
	}
}
