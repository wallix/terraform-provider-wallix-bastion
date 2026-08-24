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
			StateContext: resourceDeviceLocalDomainAccountCredentialImport,
		},
		Schema: map[string]*schema.Schema{
			skDeviceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skAccountID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{skPassword, "ssh_key"}, false),
			},
			skPassphrase: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{skPrivateKey},
			},
			skPassword: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			skPublicKey: {
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
	cfgDevice, err := readDeviceOptions(ctx, d.Get(skDeviceID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDevice.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get(skDeviceID).(string)))
	}
	cfgDomain, err := readDeviceLocalDomainOptions(ctx, d.Get(skDeviceID).(string), d.Get(skDomainID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s on device_id %s doesn't exists",
			d.Get(skDomainID).(string), d.Get(skDeviceID).(string)))
	}
	cfgAccount, err := readDeviceLocalDomainAccountOptions(ctx,
		d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgAccount.ID == "" {
		return diag.FromErr(fmt.Errorf("account_id with ID %s on domain_id %s, device_id %s doesn't exists",
			d.Get(skAccountID).(string), d.Get(skDomainID).(string), d.Get(skDeviceID).(string)))
	}
	_, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
		d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s, device_id %s already exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string), d.Get(skDeviceID).(string)))
	}
	id, err := addDeviceLocalDomainAccountCredential(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDeviceLocalDomainAccountCredential(ctx,
			d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"credential type %s on account_id %s, domain_id %s, device_id %s not found after POST",
				d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string),
				d.Get(skDeviceID).(string)))
		}
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
		d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		// WALLIX rotates a credential (manual regenerate, or an automatic
		// rotation policy) by deleting the old object and creating a new one
		// with a different internal ID, so the GET by the stored ID 404s.
		// Retry a lookup by (account, type) before treating it as deleted.
		newID, found, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
			d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string),
			d.Get(skType).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if found {
			cfg, err = readDeviceLocalDomainAccountCredentialOptions(ctx,
				d.Get(skDeviceID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), newID, m)
			if err != nil {
				return diag.FromErr(err)
			}
			if cfg.ID != "" {
				d.SetId(newID)
			}
		}
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
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
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
	if tfErr := d.Set(skDeviceID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDomainID, idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountID, idSplit[2]); tfErr != nil {
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
) (string, error) {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountCredentialJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/localdomains/"+d.Get(skDomainID).(string)+
			"/accounts/"+d.Get(skAccountID).(string)+"/credentials/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDeviceLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDeviceLocalDomainAccountCredentialJSON(d)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/localdomains/"+d.Get(skDomainID).(string)+
			"/accounts/"+d.Get(skAccountID).(string)+"/credentials/"+d.Id(), http.MethodPut, jsonData)
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
		"/devices/"+d.Get(skDeviceID).(string)+"/localdomains/"+d.Get(skDomainID).(string)+
			"/accounts/"+d.Get(skAccountID).(string)+"/credentials/"+d.Id(), http.MethodDelete, nil)
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
		Type: d.Get(skType).(string),
	}

	switch jsonData.Type {
	case skPassword:
		jsonData.Password = d.Get(skPassword).(string)
	case "ssh_key":
		jsonData.PrivateKey = d.Get(skPrivateKey).(string)
		jsonData.Passphrase = d.Get(skPassphrase).(string)
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
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPublicKey, jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
