package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
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
							Default:  "",
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							ValidateFunc: validation.StringInSlice(
								[]string{
									"SSH_SHELL_SESSION",
									"SSH_REMOTE_COMMAND",
									"SSH_SCP_UP",
									"SSH_SCP_DOWN",
									"SFTP_SESSION",
									"RLOGIN",
									"TELNET",
									"RDP",
								},
								false,
							),
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
							Default:  "",
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							Default:  "",
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							Default:  "",
						},
						"service": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
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
							Default:  "",
						},
						"application": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
		},
	}
}

func resourceTargetGroupVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_targetgroup not available with api version %s", version)
}

func resourceTargetGroupCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
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
		return diag.FromErr(fmt.Errorf("group_name %s not found after POST", d.Get("group_name").(string)))
	}
	d.SetId(id)

	return resourceTargetGroupRead(ctx, d, m)
}

func resourceTargetGroupRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
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

func resourceTargetGroupUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateTargetGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceTargetGroupRead(ctx, d, m)
}

func resourceTargetGroupDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteTargetGroup(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTargetGroupImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
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

func searchResourceTargetGroup(
	ctx context.Context, groupName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/targetgroups/?q=group_name="+groupName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonTargetGroup
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addTargetGroup(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
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
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateTargetGroup(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
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
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteTargetGroup(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/targetgroups/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareTargetGroupJSON(d *schema.ResourceData) (jsonTargetGroup, error) { //nolint: gocognit,gocyclo,maintidx
	jsonData := jsonTargetGroup{
		Description: d.Get("description").(string),
		GroupName:   d.Get("group_name").(string),
	}

	listPasswordRetrievalAccounts := d.Get("password_retrieval_accounts").(*schema.Set).List()
	jsonData.PasswordRetrieval.Accounts = make(
		[]jsonTargerGroupPasswordRetrievalAccount,
		len(listPasswordRetrievalAccounts),
	)
	for i, v := range listPasswordRetrievalAccounts {
		passwordRetrievalAccounts := v.(map[string]interface{})
		switch {
		case passwordRetrievalAccounts["domain_type"].(string) == domainTypeGlobal:
			if passwordRetrievalAccounts["device"].(string) != "" ||
				passwordRetrievalAccounts["application"].(string) != "" {
				return jsonData, errors.New("bad password_retrieval_accounts: " +
					"device and application need to be null with domain_type=global")
			}
		case passwordRetrievalAccounts["domain_type"].(string) == domainTypeLocal:
			if passwordRetrievalAccounts["device"].(string) == "" &&
				passwordRetrievalAccounts["application"].(string) == "" {
				return jsonData, errors.New("bad password_retrieval_accounts: " +
					"device or application need to be set with domain_type=local")
			}
		case passwordRetrievalAccounts["device"].(string) != "" && passwordRetrievalAccounts["application"].(string) != "":
			return jsonData, errors.New("bad password_retrieval_accounts: " +
				"device and application mutually exclusive")
		}
		jsonData.PasswordRetrieval.Accounts[i] = jsonTargerGroupPasswordRetrievalAccount{
			Account:     passwordRetrievalAccounts["account"].(string),
			Domain:      passwordRetrievalAccounts["domain"].(string),
			DomainType:  passwordRetrievalAccounts["domain_type"].(string),
			Device:      passwordRetrievalAccounts["device"].(string),
			Application: passwordRetrievalAccounts["application"].(string),
		}
	}

	listRestrictions := d.Get("restrictions").(*schema.Set).List()
	jsonData.Restrictions = make([]jsonRestriction, len(listRestrictions))
	for i, v := range listRestrictions {
		restrictions := v.(map[string]interface{})
		jsonData.Restrictions[i] = jsonRestriction{
			Action:      restrictions["action"].(string),
			Rules:       restrictions["rules"].(string),
			SubProtocol: restrictions["subprotocol"].(string),
		}
	}

	listSessionAccounts := d.Get("session_accounts").(*schema.Set).List()
	jsonData.Session.Accounts = make([]jsonTargetGroupSessionAccount, len(listSessionAccounts))
	for i, v := range listSessionAccounts {
		sessionAccounts := v.(map[string]interface{})
		switch {
		case (sessionAccounts["device"].(string) == "" || sessionAccounts["service"].(string) == "") &&
			sessionAccounts["application"].(string) == "":
			return jsonData, errors.New("bad session_accounts: " +
				"device/service or application need to be set")
		case sessionAccounts["device"].(string) != "" && sessionAccounts["application"].(string) != "":
			return jsonData, errors.New("bad session_accounts: " +
				"device and application mutually exclusive")
		case sessionAccounts["service"].(string) != "" && sessionAccounts["application"].(string) != "":
			return jsonData, errors.New("bad session_accounts: " +
				"service and application mutually exclusive")
		case sessionAccounts["device"].(string) != "" && sessionAccounts["service"].(string) == "":
			return jsonData, fmt.Errorf("bad session_accounts: "+
				"missing service for device %s", sessionAccounts["device"].(string))
		case sessionAccounts["service"].(string) != "" && sessionAccounts["device"].(string) == "":
			return jsonData, fmt.Errorf("bad session_accounts: "+
				"missing device for service %s", sessionAccounts["service"].(string))
		}
		jsonData.Session.Accounts[i] = jsonTargetGroupSessionAccount{
			Account:     sessionAccounts["account"].(string),
			Domain:      sessionAccounts["domain"].(string),
			DomainType:  sessionAccounts["domain_type"].(string),
			Device:      sessionAccounts["device"].(string),
			Service:     sessionAccounts["service"].(string),
			Application: sessionAccounts["application"].(string),
		}
	}

	listSessionAccountMappings := d.Get("session_account_mappings").(*schema.Set).List()
	jsonData.Session.AccountMappings = make([]jsonTargetGroupSessionAccountMapping, len(listSessionAccountMappings))
	for i, v := range listSessionAccountMappings {
		sessionAccountMappings := v.(map[string]interface{})
		switch {
		case sessionAccountMappings["device"].(string) != "" && sessionAccountMappings["application"].(string) != "":
			return jsonData, errors.New("bad session_account_mappings: " +
				"device and application mutually exclusive")
		case sessionAccountMappings["service"].(string) != "" && sessionAccountMappings["application"].(string) != "":
			return jsonData, errors.New("bad session_account_mappings: " +
				"service and application mutually exclusive")
		case sessionAccountMappings["device"].(string) != "" && sessionAccountMappings["service"].(string) == "":
			return jsonData, fmt.Errorf("bad session_account_mappings: "+
				"missing service for device %s", sessionAccountMappings["device"].(string))
		case sessionAccountMappings["service"].(string) != "" && sessionAccountMappings["device"].(string) == "":
			return jsonData, fmt.Errorf("bad session_account_mappings: "+
				"missing device for service %s", sessionAccountMappings["service"].(string))
		}
		jsonData.Session.AccountMappings[i] = jsonTargetGroupSessionAccountMapping{
			Device:      sessionAccountMappings["device"].(string),
			Service:     sessionAccountMappings["service"].(string),
			Application: sessionAccountMappings["application"].(string),
		}
	}

	listSessionInteractiveLogins := d.Get("session_interactive_logins").(*schema.Set).List()
	jsonData.Session.InteractiveLogins = make([]jsonTargetGroupSessionInteractiveLogin, len(listSessionInteractiveLogins))
	for i, v := range listSessionInteractiveLogins {
		sessionInteractiveLogins := v.(map[string]interface{})
		switch {
		case sessionInteractiveLogins["device"].(string) != "" && sessionInteractiveLogins["application"].(string) != "":
			return jsonData, errors.New("bad session_interactive_logins: " +
				"device and application mutually exclusive")
		case sessionInteractiveLogins["service"].(string) != "" && sessionInteractiveLogins["application"].(string) != "":
			return jsonData, errors.New("bad session_interactive_logins: " +
				"service and application mutually exclusive")
		case sessionInteractiveLogins["device"].(string) != "" && sessionInteractiveLogins["service"].(string) == "":
			return jsonData, fmt.Errorf("bad session_interactive_logins: "+
				"missing service for device %s", sessionInteractiveLogins["device"].(string))
		case sessionInteractiveLogins["service"].(string) != "" && sessionInteractiveLogins["device"].(string) == "":
			return jsonData, fmt.Errorf("bad session_interactive_logins: "+
				"missing device for service %s", sessionInteractiveLogins["service"].(string))
		}
		jsonData.Session.InteractiveLogins[i] = jsonTargetGroupSessionInteractiveLogin{
			Device:      sessionInteractiveLogins["device"].(string),
			Service:     sessionInteractiveLogins["service"].(string),
			Application: sessionInteractiveLogins["application"].(string),
		}
	}

	listSessionScenarioAccounts := d.Get("session_scenario_accounts").(*schema.Set).List()
	jsonData.Session.ScenarioAccounts = make([]jsonTargetGroupSessionScenarioAccount, len(listSessionScenarioAccounts))
	for i, v := range listSessionScenarioAccounts {
		sessionScenarioAccounts := v.(map[string]interface{})
		switch {
		case sessionScenarioAccounts["domain_type"].(string) == domainTypeGlobal:
			if sessionScenarioAccounts["device"].(string) != "" ||
				sessionScenarioAccounts["application"].(string) != "" {
				return jsonData, errors.New("bad session_scenario_accounts: " +
					"device and application need to be null with domain_type=global")
			}
		case sessionScenarioAccounts["domain_type"].(string) == domainTypeLocal:
			if sessionScenarioAccounts["device"].(string) == "" &&
				sessionScenarioAccounts["application"].(string) == "" {
				return jsonData, errors.New("bad session_scenario_accounts: " +
					"device or application need to be set with domain_type=local")
			}
		case sessionScenarioAccounts["device"].(string) != "" && sessionScenarioAccounts["application"].(string) != "":
			return jsonData, errors.New("bad session_scenario_accounts: " +
				"device and application mutually exclusive")
		}
		jsonData.Session.ScenarioAccounts[i] = jsonTargetGroupSessionScenarioAccount{
			Account:     sessionScenarioAccounts["account"].(string),
			Domain:      sessionScenarioAccounts["domain"].(string),
			DomainType:  sessionScenarioAccounts["domain_type"].(string),
			Device:      sessionScenarioAccounts["device"].(string),
			Application: sessionScenarioAccounts["application"].(string),
		}
	}

	return jsonData, nil
}

func readTargetGroupOptions(
	ctx context.Context, groupID string, m interface{},
) (
	jsonTargetGroup, error,
) {
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
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
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
	passwordRetrievalAccounts := make([]map[string]interface{}, len(jsonData.PasswordRetrieval.Accounts))
	for i, v := range jsonData.PasswordRetrieval.Accounts {
		passwordRetrievalAccounts[i] = map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"application": v.Application,
		}
	}
	if tfErr := d.Set("password_retrieval_accounts", passwordRetrievalAccounts); tfErr != nil {
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
	sessionAccounts := make([]map[string]interface{}, len(jsonData.Session.Accounts))
	for i, v := range jsonData.Session.Accounts {
		sessionAccounts[i] = map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		}
	}
	if tfErr := d.Set("session_accounts", sessionAccounts); tfErr != nil {
		panic(tfErr)
	}
	sessionAccountMappings := make([]map[string]interface{}, len(jsonData.Session.AccountMappings))
	for i, v := range jsonData.Session.AccountMappings {
		sessionAccountMappings[i] = map[string]interface{}{
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		}
	}
	if tfErr := d.Set("session_account_mappings", sessionAccountMappings); tfErr != nil {
		panic(tfErr)
	}
	sessionInteractiveLogins := make([]map[string]interface{}, len(jsonData.Session.InteractiveLogins))
	for i, v := range jsonData.Session.InteractiveLogins {
		sessionInteractiveLogins[i] = map[string]interface{}{
			"device":      v.Device,
			"service":     v.Service,
			"application": v.Application,
		}
	}
	if tfErr := d.Set("session_interactive_logins", sessionInteractiveLogins); tfErr != nil {
		panic(tfErr)
	}
	sessionScenarioAccounts := make([]map[string]interface{}, len(jsonData.Session.ScenarioAccounts))
	for i, v := range jsonData.Session.ScenarioAccounts {
		sessionScenarioAccounts[i] = map[string]interface{}{
			"account":     v.Account,
			"domain":      v.Domain,
			"domain_type": v.DomainType,
			"device":      v.Device,
			"application": v.Application,
		}
	}
	if tfErr := d.Set("session_scenario_accounts", sessionScenarioAccounts); tfErr != nil {
		panic(tfErr)
	}
}
