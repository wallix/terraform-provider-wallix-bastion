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

type jsonDeviceLocalDomain struct {
	EnablePasswordChange           bool                    `json:"enable_password_change"`
	ID                             string                  `json:"id,omitempty"`
	DomainName                     string                  `json:"domain_name"`
	AdminAccount                   *string                 `json:"admin_account,omitempty"`
	CAPrivateKey                   string                  `json:"ca_private_key,omitempty"`
	CAPublicKey                    string                  `json:"ca_public_key,omitempty"`
	Description                    string                  `json:"description"`
	Passphrase                     string                  `json:"passphrase"`
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
			State: resourceDeviceLocalDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"device_id": {
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
				RequiredWith: []string{"enable_password_change", "password_change_policy", "password_change_plugin"},
			},
			"ca_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_password_change": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"password_change_policy", "password_change_plugin"},
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
			},
		},
	}
}
func resourveDeviceLocalDomainVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_localdomain not validate with api version %s", version)
}

func resourceDeviceLocalDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDevice, err := readDeviceOptions(ctx, d.Get("device_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDevice.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get("device_id").(string)))
	}
	_, ex, err := searchResourceDeviceLocalDomain(ctx, d.Get("device_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s already exists",
			d.Get("domain_name").(string), d.Get("device_id").(string)))
	}
	err = addDeviceLocalDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomain(ctx, d.Get("device_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s can't find after POST",
			d.Get("domain_name").(string), d.Get("device_id").(string)))
	}
	d.SetId(id)

	return resourceDeviceLocalDomainRead(ctx, d, m)
}
func resourceDeviceLocalDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, d.Get("device_id").(string), d.Id(), m)
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
func resourceDeviceLocalDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceLocalDomainRead(ctx, d, m)
}
func resourceDeviceLocalDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDeviceLocalDomainImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, fmt.Errorf("id msut be <device_id>/<domain_name>")
	}
	id, ex, err := searchResourceDeviceLocalDomain(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <device_id>/<domain_name>", d.Id())
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceLocalDomain(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("device_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceLocalDomain(ctx context.Context,
	deviceID, domainName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+
		"/localdomains/?fields=domain_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDeviceLocalDomain
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

func addDeviceLocalDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainJSON(d, true)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Get("device_id").(string)+"/localdomains/",
		http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDeviceLocalDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainJSON(d, false)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDeviceLocalDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareDeviceLocalDomainJSON(d *schema.ResourceData, newResource bool) jsonDeviceLocalDomain {
	var jsonData jsonDeviceLocalDomain
	jsonData.DomainName = d.Get("domain_name").(string)
	if newResource {
		jsonData.CAPrivateKey = d.Get("ca_private_key").(string)
	} else if !strings.HasPrefix(d.Get("ca_private_key").(string), "generate:") {
		jsonData.CAPrivateKey = d.Get("ca_private_key").(string)
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
		if d.Get("password_change_plugin_parameters").(string) != "" {
			_ = json.Unmarshal([]byte(d.Get("password_change_plugin_parameters").(string)),
				&passChgPlug)
		} else {
			_ = json.Unmarshal([]byte(`{}`), &passChgPlug)
		}
		jsonData.PasswordChangePluginParameters = &passChgPlug
	}

	return jsonData
}

func readDeviceLocalDomainOptions(
	ctx context.Context, deviceID, localDomainID string, m interface{}) (jsonDeviceLocalDomain, error) {
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
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillDeviceLocalDomain(d *schema.ResourceData, jsonData jsonDeviceLocalDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
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
}
