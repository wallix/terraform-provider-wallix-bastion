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

func resourceDeviceLocalDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceLocalDomainAccountCredentialCreate,
		ReadContext:   resourceDeviceLocalDomainAccountCredentialRead,
		UpdateContext: resourceDeviceLocalDomainAccountCredentialUpdate,
		DeleteContext: resourceDeviceLocalDomainAccountCredentialDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDeviceLocalDomainAccountCredentialImport,
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
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"password", "ssh_key"}, false),
			},
			"passphrase": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{"private_key"},
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDeviceLocalDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_localdomain_account_credential "+
		"not available with api version %s", version)
}

func resourceDeviceLocalDomainAccountCredentialCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
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
	cfgAccount, err := readDeviceLocalDomainAccountOptions(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgAccount.ID == "" {
		return diag.FromErr(fmt.Errorf("account_id with ID %s on domain_id %s, device_id %s doesn't exists",
			d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	_, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), d.Get("type").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s, device_id %s already exists",
			d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	err = addDeviceLocalDomainAccountCredential(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), d.Get("type").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf(
			"credential type %s on account_id %s, domain_id %s, device_id %s not found after POST",
			d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	d.SetId(id)

	return resourceDeviceLocalDomainAccountCredentialRead(ctx, d, m)
}

func resourceDeviceLocalDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceLocalDomainAccountCredentialOptions(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDeviceLocalDomainAccountCredential(d, cfg)
	}

	return nil
}

func resourceDeviceLocalDomainAccountCredentialUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceLocalDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceLocalDomainAccountCredentialRead(ctx, d, m)
}

func resourceDeviceLocalDomainAccountCredentialDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceLocalDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDeviceLocalDomainAccountCredentialImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 4 {
		return nil, errors.New("id must be <device_id>/<domain_id>/<account_id>/<type>")
	}
	id, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx, idSplit[0], idSplit[1], idSplit[2], idSplit[3], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find credential with id %s "+
			"(id must be <device_id>/<domain_id>/<account_id>/<type>)", d.Id())
	}
	cfg, err := readDeviceLocalDomainAccountCredentialOptions(ctx, idSplit[0], idSplit[1], idSplit[2], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceLocalDomainAccountCredential(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("device_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_id", idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("account_id", idSplit[2]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceLocalDomainAccountCredential(
	ctx context.Context, deviceID, domainID, accountID, typeCred string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+deviceID+"/localdomains/"+domainID+"/accounts/"+accountID+
			"/credentials/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonCredential
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	for _, v := range results {
		if v.Type == typeCred {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDeviceLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountCredentialJSON(d)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateDeviceLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountCredentialJSON(d)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDeviceLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDeviceLocalDomainAccountCredentialJSON(
	d *schema.ResourceData,
) jsonCredential {
	jsonData := jsonCredential{
		Type: d.Get("type").(string),
	}

	if jsonData.Type == "password" {
		jsonData.Password = d.Get("password").(string)
	} else if jsonData.Type == "ssh_key" {
		jsonData.PrivateKey = d.Get("private_key").(string)
		jsonData.Passphrase = d.Get("passphrase").(string)
	}

	return jsonData
}

func readDeviceLocalDomainAccountCredentialOptions(
	ctx context.Context, deviceID, localDomainID, accountID, credentialID string, m interface{},
) (
	jsonCredential, error,
) {
	c := m.(*Client)
	var result jsonCredential
	body, code, err := c.newRequest(ctx,
		"/devices/"+deviceID+"/localdomains/"+localDomainID+
			"/accounts/"+accountID+"/credentials/"+credentialID, http.MethodGet, nil)
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
	// avoid the bug when the credential still exists but not linked to the account
	credsID, found, err := searchResourceDeviceLocalDomainAccountCredential(
		ctx, deviceID, localDomainID, accountID, result.Type, m)
	if err != nil {
		return result, err
	}
	if !found {
		return jsonCredential{}, nil
	}
	if credsID != result.ID {
		return jsonCredential{}, nil
	}

	return result, nil
}

func fillDeviceLocalDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("public_key", jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
