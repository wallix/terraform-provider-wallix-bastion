package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTargetGroupRead,
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_retrieval_accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
			"session_accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"session_account_mappings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"session_interactive_logins": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"session_scenario_accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTargetGroupVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_targetgroup not available with api version %s", version)
}

func dataSourceTargetGroupRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceTargetGroupVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceTargetGroup(ctx, d.Get("group_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("group_name %s doesn't exists", d.Get("group_name").(string)))
	}
	cfg, err := readTargetGroupOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceTargetGroup(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceTargetGroup(d *schema.ResourceData, jsonData jsonTargetGroup) {
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
