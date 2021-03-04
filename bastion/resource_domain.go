package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonDomain struct {
	EnablePasswordChange           bool                    `json:"enable_password_change"`
	ID                             string                  `json:"id,omitempty"`
	DomainName                     string                  `json:"domain_name"`
	DomainRealName                 string                  `json:"domain_real_name"`
	AdminAccount                   *string                 `json:"admin_account,omitempty"`
	CAPrivateKey                   string                  `json:"ca_private_key,omitempty"`
	CAPublicKey                    string                  `json:"ca_public_key,omitempty"`
	Description                    string                  `json:"description"`
	Passphrase                     string                  `json:"passphrase"`
	PasswordChangePolicy           string                  `json:"password_change_policy,omitempty"`
	PasswordChangePlugin           string                  `json:"password_change_plugin,omitempty"`
	PasswordChangePluginParameters *map[string]interface{} `json:"password_change_plugin_parameters,omitempty"`
	VaultPlugin                    string                  `json:"vault_plugin,omitempty"`
	VaultPluginParameters          *map[string]interface{} `json:"vault_plugin_parameters,omitempty"`
}

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_real_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_account": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enable_password_change", "password_change_policy", "password_change_plugin"},
			},
			"ca_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_private_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"vault_plugin"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_password_change": {
				Type:          schema.TypeBool,
				Optional:      true,
				RequiredWith:  []string{"password_change_policy", "password_change_plugin"},
				ConflictsWith: []string{"vault_plugin"},
			},
			"passphrase": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
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
				RequiredWith: []string{"enable_password_change", "password_change_policy", "password_change_plugin"},
				ValidateFunc: validation.StringIsJSON,
				Sensitive:    true,
			},
			"vault_plugin": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"enable_password_change", "ca_private_key"},
			},
			"vault_plugin_parameters": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"vault_plugin_parameters"},
				ValidateFunc: validation.StringIsJSON,
				Sensitive:    true,
			},
		},
	}
}
func resourceDomainVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain not validate with api version %s", version)
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceDomain(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get("domain_name").(string)))
	}
	err = addDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomain(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s can't find after POST", d.Get("domain_name").(string)))
	}
	d.SetId(id)

	return resourceDomainRead(ctx, d, m)
}
func resourceDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDomainOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDomain(d, cfg)
	}

	return nil
}
func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDomainRead(ctx, d, m)
}
func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDomainImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceDomain(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>", d.Id())
	}
	cfg, err := readDomainOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillDomain(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceDomain(ctx context.Context, domainName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/domains/?fields=domain_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDomain
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.DomainName == domainName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDomainJSON(d, true)
	body, code, err := c.newRequest(ctx, "/domains/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDomainJSON(d, false)
	body, code, err := c.newRequest(ctx, "/domains/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/domains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareDomainJSON(d *schema.ResourceData, newResource bool) jsonDomain {
	var jsonData jsonDomain
	jsonData.DomainName = d.Get("domain_name").(string)
	jsonData.DomainRealName = d.Get("domain_real_name").(string)
	if !strings.HasPrefix(d.Get("ca_private_key").(string), "generate:") {
		jsonData.CAPrivateKey = d.Get("ca_private_key").(string)
	} else if d.HasChange("ca_private_key") {
		oldKey, newKey := d.GetChange("ca_private_key")
		if oldKey.(string) == "" {
			jsonData.CAPrivateKey = newKey.(string)
		}
	}
	jsonData.Description = d.Get("description").(string)
	jsonData.Passphrase = d.Get("passphrase").(string)

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
	} else if v := d.Get("vault_plugin").(string); v != "" {
		jsonData.VaultPlugin = v
		var vaultPlugParams map[string]interface{}
		if v2 := d.Get("vault_plugin_parameters").(string); v2 != "" {
			_ = json.Unmarshal([]byte(v2),
				&vaultPlugParams)
		} else {
			_ = json.Unmarshal([]byte(`{}`), &vaultPlugParams)
		}
		jsonData.VaultPluginParameters = &vaultPlugParams
	}

	return jsonData
}

func readDomainOptions(
	ctx context.Context, domainID string, m interface{}) (jsonDomain, error) {
	c := m.(*Client)
	var result jsonDomain
	body, code, err := c.newRequest(ctx, "/domains/"+domainID, http.MethodGet, nil)
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

func fillDomain(d *schema.ResourceData, jsonData jsonDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_real_name", jsonData.DomainRealName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("admin_account", jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_public_key", jsonData.CAPublicKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable_password_change", jsonData.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("passphrase", jsonData.Passphrase); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_policy", jsonData.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_plugin", jsonData.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("vault_plugin", jsonData.VaultPlugin); tfErr != nil {
		panic(tfErr)
	}
}
