package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonApplicationLocalDomain struct {
	EnablePasswordChange           bool                    `json:"enable_password_change"`
	ID                             string                  `json:"id,omitempty"`
	AdminAccount                   *string                 `json:"admin_account,omitempty"`
	DomainName                     string                  `json:"domain_name"`
	Description                    string                  `json:"description"`
	PasswordChangePolicy           string                  `json:"password_change_policy,omitempty"`
	PasswordChangePlugin           string                  `json:"password_change_plugin,omitempty"`
	PasswordChangePluginParameters *map[string]interface{} `json:"password_change_plugin_parameters,omitempty"`
}

func resourceApplicationLocalDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationLocalDomainCreate,
		ReadContext:   resourceApplicationLocalDomainRead,
		UpdateContext: resourceApplicationLocalDomainUpdate,
		DeleteContext: resourceApplicationLocalDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceApplicationLocalDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin_account": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_password_change"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_password_change": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"password_change_policy", "password_change_plugin", "password_change_plugin_parameters"},
			},
			"password_change_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_password_change"},
			},
			"password_change_plugin": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_password_change"},
			},
			"password_change_plugin_parameters": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_password_change"},
				ValidateFunc: validation.StringIsJSON,
				Sensitive:    true,
			},
		},
	}
}

func resourceApplicationLocalDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_application_localdomain not available with api version %s", version)
}

func resourceApplicationLocalDomainCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgApplication, err := readApplicationOptions(ctx, d.Get("application_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgApplication.ID == "" {
		return diag.FromErr(fmt.Errorf("application with ID %s doesn't exists", d.Get("application_id").(string)))
	}
	_, ex, err := searchResourceApplicationLocalDomain(ctx,
		d.Get("application_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on application_id %s already exists",
			d.Get("domain_name").(string), d.Get("application_id").(string)))
	}
	err = addApplicationLocalDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplicationLocalDomain(ctx,
		d.Get("application_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on application_id %s not found after POST",
			d.Get("domain_name").(string), d.Get("application_id").(string)))
	}
	d.SetId(id)

	return resourceApplicationLocalDomainRead(ctx, d, m)
}

func resourceApplicationLocalDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readApplicationLocalDomainOptions(ctx, d.Get("application_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillApplicationLocalDomain(d, cfg)
	}

	return nil
}

func resourceApplicationLocalDomainUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateApplicationLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceApplicationLocalDomainRead(ctx, d, m)
}

func resourceApplicationLocalDomainDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteApplicationLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceApplicationLocalDomainImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("id must be <application_id>/<domain_name>")
	}
	id, ex, err := searchResourceApplicationLocalDomain(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <application_id>/<domain_name>)", d.Id())
	}
	cfg, err := readApplicationLocalDomainOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillApplicationLocalDomain(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("application_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceApplicationLocalDomain(
	ctx context.Context, applicationID, domainName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/applications/"+applicationID+
		"/localdomains/?q=domain_name="+domainName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonApplicationLocalDomain
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addApplicationLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainJSON(d, true)
	body, code, err := c.newRequest(ctx, "/applications/"+d.Get("application_id").(string)+"/localdomains/",
		http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateApplicationLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainJSON(d, false)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteApplicationLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareApplicationLocalDomainJSON(d *schema.ResourceData, newResource bool) jsonApplicationLocalDomain {
	jsonData := jsonApplicationLocalDomain{
		Description: d.Get("description").(string),
		DomainName:  d.Get("domain_name").(string),
	}

	if d.Get("enable_password_change").(bool) {
		if !newResource {
			adminAccount := d.Get("admin_account").(string)
			jsonData.AdminAccount = &adminAccount
		}
		jsonData.EnablePasswordChange = d.Get("enable_password_change").(bool)
		jsonData.PasswordChangePolicy = d.Get("password_change_policy").(string)
		jsonData.PasswordChangePlugin = d.Get("password_change_plugin").(string)
		var passChgPlug map[string]interface{}
		if v := d.Get("password_change_plugin_parameters").(string); v != "" {
			_ = json.Unmarshal([]byte(v),
				&passChgPlug)
		} else {
			_ = json.Unmarshal([]byte(`{}`), &passChgPlug)
		}
		jsonData.PasswordChangePluginParameters = &passChgPlug
	}

	return jsonData
}

func readApplicationLocalDomainOptions(
	ctx context.Context, applicationID, localDomainID string, m interface{},
) (
	jsonApplicationLocalDomain, error,
) {
	c := m.(*Client)
	var result jsonApplicationLocalDomain
	body, code, err := c.newRequest(ctx,
		"/applications/"+applicationID+"/localdomains/"+localDomainID, http.MethodGet, nil)
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

func fillApplicationLocalDomain(d *schema.ResourceData, jsonData jsonApplicationLocalDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("admin_account", jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable_password_change", jsonData.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_policy", jsonData.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_plugin", jsonData.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
}
