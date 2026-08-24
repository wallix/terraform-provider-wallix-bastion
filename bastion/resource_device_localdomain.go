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

type jsonDeviceLocalDomain struct {
	EnablePasswordChange           bool                    `json:"enable_password_change"`
	ID                             string                  `json:"id,omitempty"`
	DomainName                     string                  `json:"domain_name"`
	AdminAccount                   *string                 `json:"admin_account,omitempty"`
	CAPrivateKey                   string                  `json:"ca_private_key,omitempty"`
	CAPublicKey                    string                  `json:"ca_public_key,omitempty"`
	Description                    string                  `json:"description"`
	Passphrase                     string                  `json:"passphrase,omitempty"`
	PasswordChangePolicy           string                  `json:"password_change_policy,omitempty"`
	PasswordChangePlugin           string                  `json:"password_change_plugin,omitempty"`
	PasswordChangePluginParameters *map[string]interface{} `json:"password_change_plugin_parameters,omitempty"`
}

func resourceDeviceLocalDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceLocalDomainCreate,
		ReadContext:   resourceDeviceLocalDomainRead,
		UpdateContext: resourceDeviceLocalDomainUpdate,
		DeleteContext: resourceDeviceLocalDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDeviceLocalDomainImport,
		},
		Schema: map[string]*schema.Schema{
			skDeviceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
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
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skEnablePasswordChange: {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{skPasswordChangePolicy, skPasswordChangePlugin, skPasswordChangePluginParameters},
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
		},
	}
}

func resourceDeviceLocalDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_localdomain not available with api version %s", version)
}

func resourceDeviceLocalDomainCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDevice, err := readDeviceOptions(ctx, d.Get(skDeviceID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDevice.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get(skDeviceID).(string)))
	}
	_, ex, err := searchResourceDeviceLocalDomain(ctx, d.Get(skDeviceID).(string), d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s already exists",
			d.Get(skDomainName).(string), d.Get(skDeviceID).(string)))
	}
	id, err := addDeviceLocalDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDeviceLocalDomain(ctx, d.Get(skDeviceID).(string), d.Get(skDomainName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s not found after POST",
				d.Get(skDomainName).(string), d.Get(skDeviceID).(string)))
		}
	}
	d.SetId(id)

	return resourceDeviceLocalDomainRead(ctx, d, m)
}

func resourceDeviceLocalDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, d.Get(skDeviceID).(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDeviceLocalDomain(d, cfg)
	}

	return nil
}

func resourceDeviceLocalDomainUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceLocalDomainRead(ctx, d, m)
}

func resourceDeviceLocalDomainDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDeviceLocalDomainImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("id must be <device_id>/<domain_name>")
	}
	id, ex, err := searchResourceDeviceLocalDomain(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <device_id>/<domain_name>)", d.Id())
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceLocalDomain(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set(skDeviceID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceLocalDomain(
	ctx context.Context, deviceID, domainName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+
		"/localdomains/?q=domain_name="+domainName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonDeviceLocalDomain
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDeviceLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainJSON(d, true)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/devices/"+d.Get(skDeviceID).(string)+"/localdomains/",
		http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDeviceLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainJSON(d, false)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/localdomains/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDeviceLocalDomain(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/localdomains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDeviceLocalDomainJSON(d *schema.ResourceData, newResource bool) jsonDeviceLocalDomain {
	jsonData := jsonDeviceLocalDomain{
		Description: d.Get(skDescription).(string),
		DomainName:  d.Get(skDomainName).(string),
		Passphrase:  d.Get(skPassphrase).(string),
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
	}

	return jsonData
}

func readDeviceLocalDomainOptions(
	ctx context.Context, deviceID, localDomainID string, m interface{},
) (
	jsonDeviceLocalDomain, error,
) {
	c := m.(*Client)
	var result jsonDeviceLocalDomain
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+"/localdomains/"+localDomainID, http.MethodGet, nil)
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

func fillDeviceLocalDomain(d *schema.ResourceData, jsonData jsonDeviceLocalDomain) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
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
}
