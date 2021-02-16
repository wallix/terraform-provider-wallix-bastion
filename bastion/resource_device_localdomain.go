package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type jsonDeviceLocalDomain struct {
	EnablePasswordChange           bool   `json:"enable_password_change"`
	ID                             string `json:"id,omitempty"`
	DomainName                     string `json:"domain_name"`
	AdminAccount                   string `json:"admin_account,omitempty"`
	CAPrivateKey                   string `json:"ca_private_key,omitempty"`
	CAPublicKey                    string `json:"ca_public_key,omitempty"`
	Description                    string `json:"description"`
	Passphrase                     string `json:"passphrase"`
	PasswordChangePolicy           string `json:"password_change_policy,omitempty"`
	PasswordChangePlugin           string `json:"password_change_plugin,omitempty"`
	PasswordChangePluginParameters string `json:"password_change_plugin_parameters,omitempty"`
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
				Type:     schema.TypeString,
				Optional: true,
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
				Type:     schema.TypeBool,
				Optional: true,
			},
			"passphrase": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"password_change_policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_change_plugin": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_change_plugin_parameters": {
				Type:     schema.TypeString,
				Optional: true,
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
	c := m.(*Client)
	if err := resourveDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceLocalDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

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
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+"/localdomains/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api return not OK : %d with body %s", code, body)
	}
	var results []jsonDeviceLocalDomain
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, err
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
	json := prepareDeviceLocalDomainJSON(d, true)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Get("device_id").(string)+"/localdomains/", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
	}

	return nil
}

func updateDeviceLocalDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json := prepareDeviceLocalDomainJSON(d, false)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Id(), http.MethodPut, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
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
		return fmt.Errorf("api return not OK or NoContent : %d with body %s", code, body)
	}

	return nil
}

func prepareDeviceLocalDomainJSON(d *schema.ResourceData, newResource bool) jsonDeviceLocalDomain {
	var json jsonDeviceLocalDomain
	json.DomainName = d.Get("domain_name").(string)
	if newResource {
		json.CAPrivateKey = d.Get("ca_private_key").(string)
	} else if !strings.HasPrefix(d.Get("ca_private_key").(string), "generate:") {
		json.CAPrivateKey = d.Get("ca_private_key").(string)
	}
	json.AdminAccount = d.Get("admin_account").(string)
	json.Description = d.Get("description").(string)
	json.EnablePasswordChange = d.Get("enable_password_change").(bool)
	json.Passphrase = d.Get("passphrase").(string)
	json.PasswordChangePolicy = d.Get("password_change_policy").(string)
	json.PasswordChangePlugin = d.Get("password_change_plugin").(string)
	json.PasswordChangePluginParameters = d.Get("password_change_plugin_parameters").(string)

	return json
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
		return result, fmt.Errorf("api return not OK : %d with body %s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func fillDeviceLocalDomain(d *schema.ResourceData, json jsonDeviceLocalDomain) {
	if tfErr := d.Set("domain_name", json.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("admin_account", json.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_public_key", json.CAPublicKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", json.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable_password_change", json.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("passphrase", json.Passphrase); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_policy", json.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_plugin", json.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_plugin_parameters", json.PasswordChangePluginParameters); tfErr != nil {
		panic(tfErr)
	}
}
