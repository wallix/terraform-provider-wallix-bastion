package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

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
			StateContext: resourceTargetGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_retrieval_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						skAccount: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomain: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomainType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						skDevice: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skApplication: {
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
						skAction: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"kill", "notify"}, false),
						},
						skRules: {
							Type:     schema.TypeString,
							Required: true,
						},
						skSubprotocol: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"SSH_SHELL_SESSION",
									"SSH_REMOTE_COMMAND",
									"SSH_SCP_UP",
									"SSH_SCP_DOWN",
									"SFTP_SESSION",
									skProtoRLOGIN,
									skProtoTELNET,
									skProtoRDP,
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
						skAccount: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomain: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomainType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						skDevice: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skService: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skApplication: {
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
						skDevice: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skService: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skApplication: {
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
						skDevice: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skService: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skApplication: {
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
						skAccount: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomain: {
							Type:     schema.TypeString,
							Required: true,
						},
						skDomainType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{domainTypeLocal, domainTypeGlobal}, false),
						},
						skDevice: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						skApplication: {
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
	if slices.Contains(defaultVersionsValid(), version) {
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
	id, err := addTargetGroup(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceTargetGroup(ctx, d.Get("group_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("group_name %s not found after POST", d.Get("group_name").(string)))
		}
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
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceTargetGroup(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find group_name with id %s (id must be <group_name>)", d.Id())
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
) (string, error) {
	c := m.(*Client)
	jsonData, err := prepareTargetGroupJSON(d)
	if err != nil {
		return "", err
	}
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/targetgroups/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
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
		Description: d.Get(skDescription).(string),
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
		case passwordRetrievalAccounts[skDomainType].(string) == domainTypeGlobal:
			if passwordRetrievalAccounts[skDevice].(string) != "" ||
				passwordRetrievalAccounts[skApplication].(string) != "" {
				return jsonData, errors.New("bad password_retrieval_accounts: " +
					"device and application need to be null with domain_type=global")
			}
		case passwordRetrievalAccounts[skDomainType].(string) == domainTypeLocal:
			if passwordRetrievalAccounts[skDevice].(string) == "" &&
				passwordRetrievalAccounts[skApplication].(string) == "" {
				return jsonData, errors.New("bad password_retrieval_accounts: " +
					"device or application need to be set with domain_type=local")
			}
		case passwordRetrievalAccounts[skDevice].(string) != "" && passwordRetrievalAccounts[skApplication].(string) != "":
			return jsonData, errors.New("bad password_retrieval_accounts: " +
				"device and application mutually exclusive")
		}
		jsonData.PasswordRetrieval.Accounts[i] = jsonTargerGroupPasswordRetrievalAccount{
			Account:     passwordRetrievalAccounts[skAccount].(string),
			Domain:      passwordRetrievalAccounts[skDomain].(string),
			DomainType:  passwordRetrievalAccounts[skDomainType].(string),
			Device:      passwordRetrievalAccounts[skDevice].(string),
			Application: passwordRetrievalAccounts[skApplication].(string),
		}
	}

	listRestrictions := d.Get("restrictions").(*schema.Set).List()
	jsonData.Restrictions = make([]jsonRestriction, len(listRestrictions))
	for i, v := range listRestrictions {
		restrictions := v.(map[string]interface{})
		jsonData.Restrictions[i] = jsonRestriction{
			Action:      restrictions[skAction].(string),
			Rules:       restrictions[skRules].(string),
			SubProtocol: restrictions[skSubprotocol].(string),
		}
	}

	listSessionAccounts := d.Get("session_accounts").(*schema.Set).List()
	jsonData.Session.Accounts = make([]jsonTargetGroupSessionAccount, len(listSessionAccounts))
	for i, v := range listSessionAccounts {
		sessionAccounts := v.(map[string]interface{})
		switch {
		case (sessionAccounts[skDevice].(string) == "" || sessionAccounts[skService].(string) == "") &&
			sessionAccounts[skApplication].(string) == "":
			return jsonData, errors.New("bad session_accounts: " +
				"device/service or application need to be set")
		case sessionAccounts[skDevice].(string) != "" && sessionAccounts[skApplication].(string) != "":
			return jsonData, errors.New("bad session_accounts: " +
				"device and application mutually exclusive")
		case sessionAccounts[skService].(string) != "" && sessionAccounts[skApplication].(string) != "":
			return jsonData, errors.New("bad session_accounts: " +
				"service and application mutually exclusive")
		case sessionAccounts[skDevice].(string) != "" && sessionAccounts[skService].(string) == "":
			return jsonData, fmt.Errorf("bad session_accounts: "+
				"missing service for device %s", sessionAccounts[skDevice].(string))
		case sessionAccounts[skService].(string) != "" && sessionAccounts[skDevice].(string) == "":
			return jsonData, fmt.Errorf("bad session_accounts: "+
				"missing device for service %s", sessionAccounts[skService].(string))
		}
		jsonData.Session.Accounts[i] = jsonTargetGroupSessionAccount{
			Account:     sessionAccounts[skAccount].(string),
			Domain:      sessionAccounts[skDomain].(string),
			DomainType:  sessionAccounts[skDomainType].(string),
			Device:      sessionAccounts[skDevice].(string),
			Service:     sessionAccounts[skService].(string),
			Application: sessionAccounts[skApplication].(string),
		}
	}

	listSessionAccountMappings := d.Get("session_account_mappings").(*schema.Set).List()
	jsonData.Session.AccountMappings = make([]jsonTargetGroupSessionAccountMapping, len(listSessionAccountMappings))
	for i, v := range listSessionAccountMappings {
		sessionAccountMappings := v.(map[string]interface{})
		switch {
		case sessionAccountMappings[skDevice].(string) != "" && sessionAccountMappings[skApplication].(string) != "":
			return jsonData, errors.New("bad session_account_mappings: " +
				"device and application mutually exclusive")
		case sessionAccountMappings[skService].(string) != "" && sessionAccountMappings[skApplication].(string) != "":
			return jsonData, errors.New("bad session_account_mappings: " +
				"service and application mutually exclusive")
		case sessionAccountMappings[skDevice].(string) != "" && sessionAccountMappings[skService].(string) == "":
			return jsonData, fmt.Errorf("bad session_account_mappings: "+
				"missing service for device %s", sessionAccountMappings[skDevice].(string))
		case sessionAccountMappings[skService].(string) != "" && sessionAccountMappings[skDevice].(string) == "":
			return jsonData, fmt.Errorf("bad session_account_mappings: "+
				"missing device for service %s", sessionAccountMappings[skService].(string))
		}
		jsonData.Session.AccountMappings[i] = jsonTargetGroupSessionAccountMapping{
			Device:      sessionAccountMappings[skDevice].(string),
			Service:     sessionAccountMappings[skService].(string),
			Application: sessionAccountMappings[skApplication].(string),
		}
	}

	listSessionInteractiveLogins := d.Get("session_interactive_logins").(*schema.Set).List()
	jsonData.Session.InteractiveLogins = make([]jsonTargetGroupSessionInteractiveLogin, len(listSessionInteractiveLogins))
	for i, v := range listSessionInteractiveLogins {
		sessionInteractiveLogins := v.(map[string]interface{})
		switch {
		case sessionInteractiveLogins[skDevice].(string) != "" && sessionInteractiveLogins[skApplication].(string) != "":
			return jsonData, errors.New("bad session_interactive_logins: " +
				"device and application mutually exclusive")
		case sessionInteractiveLogins[skService].(string) != "" && sessionInteractiveLogins[skApplication].(string) != "":
			return jsonData, errors.New("bad session_interactive_logins: " +
				"service and application mutually exclusive")
		case sessionInteractiveLogins[skDevice].(string) != "" && sessionInteractiveLogins[skService].(string) == "":
			return jsonData, fmt.Errorf("bad session_interactive_logins: "+
				"missing service for device %s", sessionInteractiveLogins[skDevice].(string))
		case sessionInteractiveLogins[skService].(string) != "" && sessionInteractiveLogins[skDevice].(string) == "":
			return jsonData, fmt.Errorf("bad session_interactive_logins: "+
				"missing device for service %s", sessionInteractiveLogins[skService].(string))
		}
		jsonData.Session.InteractiveLogins[i] = jsonTargetGroupSessionInteractiveLogin{
			Device:      sessionInteractiveLogins[skDevice].(string),
			Service:     sessionInteractiveLogins[skService].(string),
			Application: sessionInteractiveLogins[skApplication].(string),
		}
	}

	listSessionScenarioAccounts := d.Get("session_scenario_accounts").(*schema.Set).List()
	jsonData.Session.ScenarioAccounts = make([]jsonTargetGroupSessionScenarioAccount, len(listSessionScenarioAccounts))
	for i, v := range listSessionScenarioAccounts {
		sessionScenarioAccounts := v.(map[string]interface{})
		switch {
		case sessionScenarioAccounts[skDomainType].(string) == domainTypeGlobal:
			if sessionScenarioAccounts[skDevice].(string) != "" ||
				sessionScenarioAccounts[skApplication].(string) != "" {
				return jsonData, errors.New("bad session_scenario_accounts: " +
					"device and application need to be null with domain_type=global")
			}
		case sessionScenarioAccounts[skDomainType].(string) == domainTypeLocal:
			if sessionScenarioAccounts[skDevice].(string) == "" &&
				sessionScenarioAccounts[skApplication].(string) == "" {
				return jsonData, errors.New("bad session_scenario_accounts: " +
					"device or application need to be set with domain_type=local")
			}
		case sessionScenarioAccounts[skDevice].(string) != "" && sessionScenarioAccounts[skApplication].(string) != "":
			return jsonData, errors.New("bad session_scenario_accounts: " +
				"device and application mutually exclusive")
		}
		jsonData.Session.ScenarioAccounts[i] = jsonTargetGroupSessionScenarioAccount{
			Account:     sessionScenarioAccounts[skAccount].(string),
			Domain:      sessionScenarioAccounts[skDomain].(string),
			DomainType:  sessionScenarioAccounts[skDomainType].(string),
			Device:      sessionScenarioAccounts[skDevice].(string),
			Application: sessionScenarioAccounts[skApplication].(string),
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
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	passwordRetrievalAccounts := make([]map[string]interface{}, len(jsonData.PasswordRetrieval.Accounts))
	for i, v := range jsonData.PasswordRetrieval.Accounts {
		passwordRetrievalAccounts[i] = map[string]interface{}{
			skAccount:     v.Account,
			skDomain:      v.Domain,
			skDomainType:  v.DomainType,
			skDevice:      v.Device,
			skApplication: v.Application,
		}
	}
	if tfErr := d.Set("password_retrieval_accounts", passwordRetrievalAccounts); tfErr != nil {
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
	sessionAccounts := make([]map[string]interface{}, len(jsonData.Session.Accounts))
	for i, v := range jsonData.Session.Accounts {
		sessionAccounts[i] = map[string]interface{}{
			skAccount:     v.Account,
			skDomain:      v.Domain,
			skDomainType:  v.DomainType,
			skDevice:      v.Device,
			skService:     v.Service,
			skApplication: v.Application,
		}
	}
	if tfErr := d.Set("session_accounts", sessionAccounts); tfErr != nil {
		panic(tfErr)
	}
	sessionAccountMappings := make([]map[string]interface{}, len(jsonData.Session.AccountMappings))
	for i, v := range jsonData.Session.AccountMappings {
		sessionAccountMappings[i] = map[string]interface{}{
			skDevice:      v.Device,
			skService:     v.Service,
			skApplication: v.Application,
		}
	}
	if tfErr := d.Set("session_account_mappings", sessionAccountMappings); tfErr != nil {
		panic(tfErr)
	}
	sessionInteractiveLogins := make([]map[string]interface{}, len(jsonData.Session.InteractiveLogins))
	for i, v := range jsonData.Session.InteractiveLogins {
		sessionInteractiveLogins[i] = map[string]interface{}{
			skDevice:      v.Device,
			skService:     v.Service,
			skApplication: v.Application,
		}
	}
	if tfErr := d.Set("session_interactive_logins", sessionInteractiveLogins); tfErr != nil {
		panic(tfErr)
	}
	sessionScenarioAccounts := make([]map[string]interface{}, len(jsonData.Session.ScenarioAccounts))
	for i, v := range jsonData.Session.ScenarioAccounts {
		sessionScenarioAccounts[i] = map[string]interface{}{
			skAccount:     v.Account,
			skDomain:      v.Domain,
			skDomainType:  v.DomainType,
			skDevice:      v.Device,
			skApplication: v.Application,
		}
	}
	if tfErr := d.Set("session_scenario_accounts", sessionScenarioAccounts); tfErr != nil {
		panic(tfErr)
	}
}
