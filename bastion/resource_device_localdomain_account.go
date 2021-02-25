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

type jsonDeviceLocalDomainAccount struct {
	ID                   string            `json:"id,omitempty"`
	AccountName          string            `json:"account_name"`
	AccountLogin         string            `json:"account_login"`
	Description          string            `json:"description"`
	DomainPasswordChange *bool             `json:"domain_password_change,omitempty"`
	AutoChangePassword   bool              `json:"auto_change_password"`
	AutoChangeSSHKey     bool              `json:"auto_change_ssh_key"`
	CheckoutPolicy       string            `json:"checkout_policy"`
	CertificateValidity  string            `json:"certificate_validity,omitempty"`
	Services             []string          `json:"services"`
	Credentials          *[]jsonCredential `json:"credentials,omitempty"`
}

func resourceDeviceLocalDomainAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceLocalDomainAccountCreate,
		ReadContext:   resourceDeviceLocalDomainAccountRead,
		UpdateContext: resourceDeviceLocalDomainAccountUpdate,
		DeleteContext: resourceDeviceLocalDomainAccountDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDeviceLocalDomainAccountImport,
		},
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_login": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auto_change_password": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_change_ssh_key": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"certificate_validity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"checkout_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourveDeviceLocalDomainAccountVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_localdomain_account not validate with api version %s", version)
}

func resourceDeviceLocalDomainAccountCreate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDevice, err := readDeviceOptions(ctx, d.Get("device_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDevice.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get("device_id").(string)))
	}
	cfgDomain, err := readDeviceLocalDomainOptions(ctx, d.Get("device_id").(string), d.Get("domain_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s on device_id %s doesn't exists",
			d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	_, ex, err := searchResourceDeviceLocalDomainAccount(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, device_id %s already exists",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	err = addDeviceLocalDomainAccount(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomainAccount(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, device_id %s can't find after POST",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	d.SetId(id)

	return resourceDeviceLocalDomainAccountRead(ctx, d, m)
}
func resourceDeviceLocalDomainAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceLocalDomainAccountOptions(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDeviceLocalDomainAccount(d, cfg)
	}

	return nil
}
func resourceDeviceLocalDomainAccountUpdate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceLocalDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceLocalDomainAccountRead(ctx, d, m)
}
func resourceDeviceLocalDomainAccountDelete(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceLocalDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDeviceLocalDomainAccountImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return nil, fmt.Errorf("id must be <device_id>/<domain_id>/<account_name>")
	}
	id, ex, err := searchResourceDeviceLocalDomainAccount(ctx, idSplit[0], idSplit[1], idSplit[2], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find account_name with id %s "+
			"(id must be <device_id>/<domain_id>/<account_name>", d.Id())
	}
	cfg, err := readDeviceLocalDomainAccountOptions(ctx, idSplit[0], idSplit[1], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceLocalDomainAccount(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("device_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_id", idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceLocalDomainAccount(ctx context.Context,
	deviceID, domainID, accountName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+"/localdomains/"+domainID+
		"/accounts/?fields=account_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDeviceLocalDomainAccount
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.AccountName == accountName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDeviceLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountJSON(d)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDeviceLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountJSON(d)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDeviceLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareDeviceLocalDomainAccountJSON(d *schema.ResourceData) jsonDeviceLocalDomainAccount {
	var jsonData jsonDeviceLocalDomainAccount
	jsonData.AccountName = d.Get("account_name").(string)
	jsonData.AccountLogin = d.Get("account_login").(string)
	jsonData.CheckoutPolicy = d.Get("checkout_policy").(string)
	jsonData.AutoChangePassword = d.Get("auto_change_password").(bool)
	jsonData.AutoChangeSSHKey = d.Get("auto_change_ssh_key").(bool)
	jsonData.CertificateValidity = d.Get("certificate_validity").(string)
	jsonData.Description = d.Get("description").(string)
	if len(d.Get("services").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("services").(*schema.Set).List() {
			jsonData.Services = append(jsonData.Services, v.(string))
		}
	} else {
		jsonData.Services = make([]string, 0)
	}

	return jsonData
}

func readDeviceLocalDomainAccountOptions(
	ctx context.Context, deviceID, localDomainID, accountID string, m interface{}) (
	jsonDeviceLocalDomainAccount, error) {
	c := m.(*Client)
	var result jsonDeviceLocalDomainAccount
	body, code, err := c.newRequest(ctx,
		"/devices/"+deviceID+"/localdomains/"+localDomainID+
			"/accounts/"+accountID, http.MethodGet, nil)
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

func fillDeviceLocalDomainAccount(d *schema.ResourceData, jsonData jsonDeviceLocalDomainAccount) {
	if tfErr := d.Set("account_name", jsonData.AccountName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("account_login", jsonData.AccountLogin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("checkout_policy", jsonData.CheckoutPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_password", jsonData.AutoChangePassword); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_ssh_key", jsonData.AutoChangeSSHKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_validity", jsonData.CertificateValidity); tfErr != nil {
		panic(tfErr)
	}
	credentials := make([]map[string]interface{}, 0)
	for _, v := range *jsonData.Credentials {
		credentials = append(credentials, map[string]interface{}{
			"id":         v.ID,
			"public_key": v.PublicKey,
			"type":       v.Type,
		})
	}
	if tfErr := d.Set("credentials", credentials); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_password_change", jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("services", jsonData.Services); tfErr != nil {
		panic(tfErr)
	}
}
