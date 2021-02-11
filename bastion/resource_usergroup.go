package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonUserGroup struct {
	Users        *[]string                    `json:"users,omitempty"`
	ID           string                       `json:"id,omitempty"`
	Description  string                       `json:"description"`
	GroupName    string                       `json:"group_name"`
	Profile      string                       `json:"profile"`
	TimeFrames   []string                     `json:"timeframes"`
	Restrictions *[]jsonUserGroupRestrictions `json:"restrictions,omitempty"`
}
type jsonUserGroupRestrictions struct {
	Action      string `json:"action"`
	Rules       string `json:"rules"`
	SubProtocol string `json:"subprotocol"`
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
				Type:     schema.TypeList,
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
				Type:     schema.TypeList,
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
func resourveUserGroupVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_usergroup not validate with api version %v", version)
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceUserGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("group_name %v already exists", d.Get("group_name").(string)))
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
		return diag.FromErr(fmt.Errorf("group_name %v can't find after POST", d.Get("group_name").(string)))
	}
	d.SetId(id)

	return resourceUserGroupRead(ctx, d, m)
}
func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	config, err := readUserGroupOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if config.ID == "" {
		d.SetId("")
	} else {
		fillUserGroup(d, config)
	}

	return nil
}
func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateUserGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceUserGroupRead(ctx, d, m)
}
func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
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
	if err := resourveUserGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceUserGroup(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find group_name with id %v (id must be <group_name>", d.Id())
	}
	config, err := readUserGroupOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillUserGroup(d, config)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceUserGroup(ctx context.Context, groupName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/usergroups/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api return not OK : %d with body %s", code, body)
	}
	var results []jsonUserGroup
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, err
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
	json := prepareUserGroupJSON(d, true)
	body, code, err := c.newRequest(ctx, "/usergroups/", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
	}

	return nil
}

func updateUserGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json := prepareUserGroupJSON(d, false)
	body, code, err := c.newRequest(ctx, "/usergroups/"+d.Id()+"?force=true", http.MethodPut, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
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
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
	}

	return nil
}

func prepareUserGroupJSON(d *schema.ResourceData, newResource bool) jsonUserGroup {
	group := jsonUserGroup{
		Description: d.Get("description").(string),
		GroupName:   d.Get("group_name").(string),
		Profile:     d.Get("profile").(string),
	}
	if newResource {
		if d.Get("users") != nil {
			users := make([]string, 0)
			for _, v := range d.Get("users").(*schema.Set).List() {
				users = append(users, v.(string))
			}
			group.Users = &users
		}
	}
	if d.HasChanges("users") {
		users := make([]string, 0)
		for _, v := range d.Get("users").(*schema.Set).List() {
			users = append(users, v.(string))
		}
		group.Users = &users
	}
	for _, v := range d.Get("timeframes").([]interface{}) {
		group.TimeFrames = append(group.TimeFrames, v.(string))
	}
	if len(d.Get("restrictions").([]interface{})) > 0 {
		groupRestrictions := make([]jsonUserGroupRestrictions, 0)
		for _, v := range d.Get("restrictions").([]interface{}) {
			r := v.(map[string]interface{})
			groupRestrictions = append(groupRestrictions, jsonUserGroupRestrictions{
				Action:      r["action"].(string),
				Rules:       r["rules"].(string),
				SubProtocol: r["subprotocol"].(string),
			})
		}
		group.Restrictions = &groupRestrictions
	}

	return group
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
		return result, fmt.Errorf("api return not OK : %d with body %s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func fillUserGroup(d *schema.ResourceData, json jsonUserGroup) {
	if tfErr := d.Set("group_name", json.GroupName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("timeframes", json.TimeFrames); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", json.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("profile", json.Profile); tfErr != nil {
		panic(tfErr)
	}
	restrictions := make([]map[string]interface{}, 0)
	for _, v := range *json.Restrictions {
		restrictions = append(restrictions, map[string]interface{}{
			"action":      v.Action,
			"rules":       v.Rules,
			"subprotocol": v.SubProtocol,
		})
	}
	if tfErr := d.Set("restrictions", restrictions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("users", json.Users); tfErr != nil {
		panic(tfErr)
	}
}
