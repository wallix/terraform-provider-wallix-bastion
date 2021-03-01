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

const (
	domainTypeGlobal = "global"
	domainTypeLocal  = "local"
)

type jsonTargetGroup struct {
	ID                string                           `json:"id,omitempty"`
	Description       string                           `json:"description"`
	GroupName         string                           `json:"group_name"`
	PasswordRetrieval jsonTargerGroupPasswordRetrieval `json:"password_retrieval"`
	Restrictions      []jsonRestriction                `json:"restrictions"`
	Session           jsonTargetGroupSession           `json:"session"`
}

type jsonTargerGroupPasswordRetrieval struct {
	Accounts []jsonTargerGroupPasswordRetrievalAccount `json:"accounts"`
}
type jsonTargetGroupSession struct {
	Accounts          []jsonTargetGroupSessionAccount          `json:"accounts"`
	AccountMappings   []jsonTargetGroupSessionAccountMapping   `json:"account_mappings"`
	InteractiveLogins []jsonTargetGroupSessionInteractiveLogin `json:"interactive_logins"`
	ScenarioAccounts  []jsonTargetGroupSessionScenarioAccount  `json:"scenario_accounts"`
}

type jsonTargerGroupPasswordRetrievalAccount struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	DomainType  string `json:"domain_type"`
	Device      string `json:"device"`
	Application string `json:"application"`
}
type jsonTargetGroupSessionAccount struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	DomainType  string `json:"domain_type"`
	Device      string `json:"device"`
	Service     string `json:"service"`
	Application string `json:"application"`
}
type jsonTargetGroupSessionAccountMapping struct {
	Device      string `json:"device"`
	Service     string `json:"service"`
	Application string `json:"application"`
}
type jsonTargetGroupSessionInteractiveLogin struct {
	Device      string `json:"device"`
	Service     string `json:"service"`
	Application string `json:"application"`
}
type jsonTargetGroupSessionScenarioAccount struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	DomainType  string `json:"domain_type"`
	Device      string `json:"device"`
	Application string `json:"application"`
}

func resourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTargetGroupCreate,
		ReadContext:   resourceTargetGroupRead,
		UpdateContext: resourceTargetGroupUpdate,
		DeleteContext: resourceTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTargetGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_retrieval_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						"device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
			"session_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						"device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"session_account_mappings": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"session_interactive_logins": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"session_scenario_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"domain_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						"device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
func resourveTargetGroupVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_targetgroup not validate with api version %s", version)
}

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceTargetGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("group_name %s already exists", d.Get("group_name").(string)))
	}
	err = addTargetGroup(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceTargetGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("group_name %s can't find after POST", d.Get("group_name").(string)))
	}
	d.SetId(id)

	return resourceTargetGroupRead(ctx, d, m)
}
func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readTargetGroupOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillTargetGroup(d, cfg)
	}

	return nil
}
func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateTargetGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceTargetGroupRead(ctx, d, m)
}
func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteTargetGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceTargetGroupImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceTargetGroup(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find group_name with id %s (id must be <group_name>", d.Id())
	}
	cfg, err := readTargetGroupOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillTargetGroup(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceTargetGroup(ctx context.Context, groupName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/targetgroups/?fields=group_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonTargetGroup
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

func addTargetGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json, err := prepareTargetGroupJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/targetgroups/", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateTargetGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json, err := prepareTargetGroupJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/targetgroups/"+d.Id()+"?force=true", http.MethodPut, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteTargetGroup(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/targetgroups/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareTargetGroupJSON(d *schema.ResourceData) (jsonTargetGroup, error) { // nolint: gocognit, gocyclo
	jsonData := jsonTargetGroup{
		Description: d.Get("description").(string),
		GroupName:   d.Get("group_name").(string),
	}
	if len(d.Get("password_retrieval_accounts").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("password_retrieval_accounts").(*schema.Set).List() {
			passRetrievalAccounts := v.(map[string]interface{})
			switch {
			case passRetrievalAccounts["domain_type"].(string) == domainTypeGlobal:
				if passRetrievalAccounts["device"].(string) != "" ||
					passRetrievalAccounts["application"].(string) != "" {
					return jsonData, fmt.Errorf("bad password_retrieval_accounts: " +
						"device and application need to be null with domain_type=global")
				}
			case passRetrievalAccounts["domain_type"].(string) == domainTypeLocal:
				if passRetrievalAccounts["device"].(string) == "" &&
					passRetrievalAccounts["application"].(string) == "" {
					return jsonData, fmt.Errorf("bad password_retrieval_accounts: " +
						"device or application need to be set with domain_type=local")
				}
			case passRetrievalAccounts["device"].(string) != "" && passRetrievalAccounts["application"].(string) != "":
				return jsonData, fmt.Errorf("bad password_retrieval_accounts: " +
					"device and application mutually exclusive")
			}
			jsonData.PasswordRetrieval.Accounts = append(jsonData.PasswordRetrieval.Accounts,
				jsonTargerGroupPasswordRetrievalAccount{
					Account:     passRetrievalAccounts["account"].(string),
					Domain:      passRetrievalAccounts["domain"].(string),
					DomainType:  passRetrievalAccounts["domain_type"].(string),
					Device:      passRetrievalAccounts["device"].(string),
					Application: passRetrievalAccounts["application"].(string),
				})
		}
	} else {
		jsonData.PasswordRetrieval.Accounts = make([]jsonTargerGroupPasswordRetrievalAccount, 0)
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
	if len(d.Get("session_accounts").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("session_accounts").(*schema.Set).List() {
			sessAccounts := v.(map[string]interface{})
			switch {
			case sessAccounts["domain_type"].(string) == domainTypeGlobal:
				if sessAccounts["device"].(string) != "" ||
					sessAccounts["service"].(string) != "" ||
					sessAccounts["application"].(string) != "" {
					return jsonData, fmt.Errorf("bad session_accounts: " +
						"device,service,application need to be null with domain_type=global")
				}
			case sessAccounts["domain_type"].(string) == domainTypeLocal:
				if (sessAccounts["device"].(string) == "" ||
					sessAccounts["service"].(string) == "") &&
					sessAccounts["application"].(string) == "" {
					return jsonData, fmt.Errorf("bad session_accounts: " +
						"device/service or application need to be set with domain_type=local")
				}
			case sessAccounts["device"].(string) != "" && sessAccounts["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_accounts: " +
					"device and application mutually exclusive")
			case sessAccounts["service"].(string) != "" && sessAccounts["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_accounts: " +
					"service and application mutually exclusive")
			case sessAccounts["device"].(string) != "" && sessAccounts["service"].(string) == "":
				return jsonData, fmt.Errorf("bad session_accounts: "+
					"missing service for device %s", sessAccounts["device"].(string))
			case sessAccounts["service"].(string) != "" && sessAccounts["device"].(string) == "":
				return jsonData, fmt.Errorf("bad session_accounts: "+
					"missing device for service %s", sessAccounts["service"].(string))
			}
			jsonData.Session.Accounts = append(jsonData.Session.Accounts,
				jsonTargetGroupSessionAccount{
					Account:     sessAccounts["account"].(string),
					Domain:      sessAccounts["domain"].(string),
					DomainType:  sessAccounts["domain_type"].(string),
					Device:      sessAccounts["device"].(string),
					Service:     sessAccounts["service"].(string),
					Application: sessAccounts["application"].(string),
				})
		}
	} else {
		jsonData.Session.Accounts = make([]jsonTargetGroupSessionAccount, 0)
	}
	if len(d.Get("session_account_mappings").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("session_account_mappings").(*schema.Set).List() {
			sessAccountMappings := v.(map[string]interface{})
			switch {
			case sessAccountMappings["device"].(string) != "" && sessAccountMappings["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_account_mappings: " +
					"device and application mutually exclusive")
			case sessAccountMappings["service"].(string) != "" && sessAccountMappings["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_account_mappings: " +
					"service and application mutually exclusive")
			case sessAccountMappings["device"].(string) != "" && sessAccountMappings["service"].(string) == "":
				return jsonData, fmt.Errorf("bad session_account_mappings: "+
					"missing service for device %s", sessAccountMappings["device"].(string))
			case sessAccountMappings["service"].(string) != "" && sessAccountMappings["device"].(string) == "":
				return jsonData, fmt.Errorf("bad session_account_mappings: "+
					"missing device for service %s", sessAccountMappings["service"].(string))
			}
			jsonData.Session.AccountMappings = append(jsonData.Session.AccountMappings,
				jsonTargetGroupSessionAccountMapping{
					Device:      sessAccountMappings["device"].(string),
					Service:     sessAccountMappings["service"].(string),
					Application: sessAccountMappings["application"].(string),
				})
		}
	} else {
		jsonData.Session.AccountMappings = make([]jsonTargetGroupSessionAccountMapping, 0)
	}
	if len(d.Get("session_interactive_logins").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("session_interactive_logins").(*schema.Set).List() {
			sessInteractiveLogins := v.(map[string]interface{})
			switch {
			case sessInteractiveLogins["device"].(string) != "" && sessInteractiveLogins["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_interactive_logins: " +
					"device and application mutually exclusive")
			case sessInteractiveLogins["service"].(string) != "" && sessInteractiveLogins["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_interactive_logins: " +
					"service and application mutually exclusive")
			case sessInteractiveLogins["device"].(string) != "" && sessInteractiveLogins["service"].(string) == "":
				return jsonData, fmt.Errorf("bad session_interactive_logins: "+
					"missing service for device %s", sessInteractiveLogins["device"].(string))
			case sessInteractiveLogins["service"].(string) != "" && sessInteractiveLogins["device"].(string) == "":
				return jsonData, fmt.Errorf("bad session_interactive_logins: "+
					"missing device for service %s", sessInteractiveLogins["service"].(string))
			}
			jsonData.Session.InteractiveLogins = append(jsonData.Session.InteractiveLogins,
				jsonTargetGroupSessionInteractiveLogin{
					Device:      sessInteractiveLogins["device"].(string),
					Service:     sessInteractiveLogins["service"].(string),
					Application: sessInteractiveLogins["application"].(string),
				})
		}
	} else {
		jsonData.Session.InteractiveLogins = make([]jsonTargetGroupSessionInteractiveLogin, 0)
	}
	if len(d.Get("session_scenario_accounts").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("session_scenario_accounts").(*schema.Set).List() {
			sessScenarioAccounts := v.(map[string]interface{})
			switch {
			case sessScenarioAccounts["domain_type"].(string) == domainTypeGlobal:
				if sessScenarioAccounts["device"].(string) != "" ||
					sessScenarioAccounts["application"].(string) != "" {
					return jsonData, fmt.Errorf("bad session_scenario_accounts: " +
						"device and application need to be null with domain_type=global")
				}
			case sessScenarioAccounts["domain_type"].(string) == domainTypeLocal:
				if sessScenarioAccounts["device"].(string) == "" &&
					sessScenarioAccounts["application"].(string) == "" {
					return jsonData, fmt.Errorf("bad session_scenario_accounts: " +
						"device or application need to be set with domain_type=local")
				}
			case sessScenarioAccounts["device"].(string) != "" && sessScenarioAccounts["application"].(string) != "":
				return jsonData, fmt.Errorf("bad session_scenario_accounts: " +
					"device and application mutually exclusive")
			}
			jsonData.Session.ScenarioAccounts = append(jsonData.Session.ScenarioAccounts,
				jsonTargetGroupSessionScenarioAccount{
					Account:     sessScenarioAccounts["account"].(string),
					Domain:      sessScenarioAccounts["domain"].(string),
					DomainType:  sessScenarioAccounts["domain_type"].(string),
					Device:      sessScenarioAccounts["device"].(string),
					Application: sessScenarioAccounts["application"].(string),
				})
		}
	} else {
		jsonData.Session.ScenarioAccounts = make([]jsonTargetGroupSessionScenarioAccount, 0)
	}

	return jsonData, nil
}

func readTargetGroupOptions(
	ctx context.Context, groupID string, m interface{}) (jsonTargetGroup, error) {
	c := m.(*Client)
	var result jsonTargetGroup
	body, code, err := c.newRequest(ctx, "/targetgroups/"+groupID, http.MethodGet, nil)
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

func fillTargetGroup(d *schema.ResourceData, jsonData jsonTargetGroup) {
	if tfErr := d.Set("group_name", jsonData.GroupName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	passwordRetrievalAccounts := make([]map[string]interface{}, 0)
	for _, v := range jsonData.PasswordRetrieval.Accounts {
		passwordRetrievalAccounts = append(passwordRetrievalAccounts, map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"application": v.Application,
		})
	}
	if tfErr := d.Set("password_retrieval_accounts", passwordRetrievalAccounts); tfErr != nil {
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
	sessionAccounts := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Session.Accounts {
		sessionAccounts = append(sessionAccounts, map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		})
	}
	if tfErr := d.Set("session_accounts", sessionAccounts); tfErr != nil {
		panic(tfErr)
	}
	sessionAccountsMappings := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Session.AccountMappings {
		sessionAccountsMappings = append(sessionAccountsMappings, map[string]interface{}{
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		})
	}
	if tfErr := d.Set("session_account_mappings", sessionAccountsMappings); tfErr != nil {
		panic(tfErr)
	}
	sessionInteractiveLogins := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Session.InteractiveLogins {
		sessionInteractiveLogins = append(sessionInteractiveLogins, map[string]interface{}{
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		})
	}
	if tfErr := d.Set("session_interactive_logins", sessionInteractiveLogins); tfErr != nil {
		panic(tfErr)
	}
	sessionScenarioAccounts := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Session.ScenarioAccounts {
		sessionScenarioAccounts = append(sessionScenarioAccounts, map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"application": v.Application,
		})
	}
	if tfErr := d.Set("session_scenario_accounts", sessionScenarioAccounts); tfErr != nil {
		panic(tfErr)
	}
}
