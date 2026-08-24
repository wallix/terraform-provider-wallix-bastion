package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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
	Passphrase                     string                  `json:"passphrase,omitempty"`
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
			StateContext: resourceDomainImport,
		},
		Schema: map[string]*schema.Schema{
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_real_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			skAdminAccount: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{skEnablePasswordChange, skPasswordChangePolicy, skPasswordChangePlugin},
			},
			skCAPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCAPrivateKey: {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{skVaultPlugin},
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skEnablePasswordChange: {
				Type:          schema.TypeBool,
				Optional:      true,
				RequiredWith:  []string{skPasswordChangePolicy, skPasswordChangePlugin, skPasswordChangePluginParameters},
				ConflictsWith: []string{skVaultPlugin},
			},
			skPassphrase: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{skCAPrivateKey},
			},
			skPasswordChangePolicy: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{skEnablePasswordChange},
			},
			skPasswordChangePlugin: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{skEnablePasswordChange},
			},
			skPasswordChangePluginParameters: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{skEnablePasswordChange},
				ValidateFunc: validation.StringIsJSON,
				Sensitive:    true,
			},
			skVaultPlugin: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{skEnablePasswordChange, skCAPrivateKey},
				RequiredWith:  []string{"vault_plugin_parameters"},
			},
			"vault_plugin_parameters": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{skVaultPlugin},
				ValidateFunc: validation.StringIsJSON,
				Sensitive:    true,
			},
		},
	}
}

func resourceDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain not available with api version %s", version)
}

func resourceDomainCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceDomain(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get(skDomainName).(string)))
	}
	id, err := addDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDomain(ctx, d.Get(skDomainName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("domain_name %s not found after POST", d.Get(skDomainName).(string)))
		}
	}
	d.SetId(id)

	return resourceDomainRead(ctx, d, m)
}

func resourceDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
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

func resourceDomainUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
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

func resourceDomainDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceDomain(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>)", d.Id())
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

func searchResourceDomain(
	ctx context.Context, domainName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/domains/?q=domain_name="+domainName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonDomain
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareDomainJSON(d, true)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/domains/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDomainJSON(d, false)
	body, code, err := c.newRequest(ctx, "/domains/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/domains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDomainJSON(d *schema.ResourceData, newResource bool) jsonDomain {
	jsonData := jsonDomain{
		Description:    d.Get(skDescription).(string),
		DomainName:     d.Get(skDomainName).(string),
		DomainRealName: d.Get("domain_real_name").(string),
		Passphrase:     d.Get(skPassphrase).(string),
	}

	if !strings.HasPrefix(d.Get(skCAPrivateKey).(string), "generate:") {
		jsonData.CAPrivateKey = d.Get(skCAPrivateKey).(string)
	} else if d.HasChange(skCAPrivateKey) {
		oldKey, newKey := d.GetChange(skCAPrivateKey)
		if oldKey.(string) == "" {
			jsonData.CAPrivateKey = newKey.(string)
		}
	}

	if d.Get(skEnablePasswordChange).(bool) {
		if !newResource {
			adminAccount := d.Get(skAdminAccount).(string)
			jsonData.AdminAccount = &adminAccount
		}
		jsonData.EnablePasswordChange = d.Get(skEnablePasswordChange).(bool)
		jsonData.PasswordChangePolicy = d.Get(skPasswordChangePolicy).(string)
		jsonData.PasswordChangePlugin = d.Get(skPasswordChangePlugin).(string)
		var passChgPlug map[string]interface{}
		if v := d.Get(skPasswordChangePluginParameters).(string); v != "" {
			_ = json.Unmarshal([]byte(v),
				&passChgPlug)
		} else {
			_ = json.Unmarshal([]byte(`{}`), &passChgPlug)
		}
		jsonData.PasswordChangePluginParameters = &passChgPlug
	} else if v := d.Get(skVaultPlugin).(string); v != "" {
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
	ctx context.Context, domainID string, m interface{},
) (
	jsonDomain, error,
) {
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
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}

func fillDomain(d *schema.ResourceData, jsonData jsonDomain) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_real_name", jsonData.DomainRealName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAdminAccount, jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCAPublicKey, jsonData.CAPublicKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skEnablePasswordChange, jsonData.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPasswordChangePolicy, jsonData.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPasswordChangePlugin, jsonData.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skVaultPlugin, jsonData.VaultPlugin); tfErr != nil {
		panic(tfErr)
	}
}
