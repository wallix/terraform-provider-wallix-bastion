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

func resourceApplicationLocalDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationLocalDomainAccountCredentialCreate,
		ReadContext:   resourceApplicationLocalDomainAccountCredentialRead,
		UpdateContext: resourceApplicationLocalDomainAccountCredentialUpdate,
		DeleteContext: resourceApplicationLocalDomainAccountCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationLocalDomainAccountCredentialImport,
		},
		Schema: map[string]*schema.Schema{
			"application_id": {
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
				ValidateFunc: validation.StringInSlice([]string{"password"}, false),
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceApplicationLocalDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_application_localdomain_account_credential "+
		"not available with api version %s", version)
}

func resourceApplicationLocalDomainAccountCredentialCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgApplication, err := readApplicationOptions(ctx, d.Get("application_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgApplication.ID == "" {
		return diag.FromErr(fmt.Errorf("application with ID %s doesn't exists", d.Get("application_id").(string)))
	}
	cfgDomain, err := readApplicationLocalDomainOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s on application_id %s doesn't exists",
			d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	cfgAccount, err := readApplicationLocalDomainAccountOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgAccount.ID == "" {
		return diag.FromErr(fmt.Errorf("account_id with ID %s on domain_id %s, application_id %s doesn't exists",
			d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	_, ex, err := searchResourceApplicationLocalDomainAccountCredential(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string),
		d.Get("type").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf(
			"credential type %s on account_id %s, domain_id %s, application_id %s already exists",
			d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	id, err := addApplicationLocalDomainAccountCredential(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceApplicationLocalDomainAccountCredential(ctx,
			d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string),
			d.Get("type").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"credential type %s on account_id %s, domain_id %s, application_id %s not found after POST",
				d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string),
				d.Get("application_id").(string)))
		}
	}
	d.SetId(id)

	return resourceApplicationLocalDomainAccountCredentialRead(ctx, d, m)
}

func resourceApplicationLocalDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readApplicationLocalDomainAccountCredentialOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		// WALLIX rotates a credential (manual regenerate, or an automatic
		// rotation policy) by deleting the old object and creating a new one
		// with a different internal ID, so the GET by the stored ID 404s.
		// Retry a lookup by (account, type) before treating it as deleted.
		newID, found, err := searchResourceApplicationLocalDomainAccountCredential(ctx,
			d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string),
			d.Get("type").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if found {
			cfg, err = readApplicationLocalDomainAccountCredentialOptions(ctx,
				d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), newID, m)
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
		fillApplicationLocalDomainAccountCredential(d, cfg)
	}

	return nil
}

func resourceApplicationLocalDomainAccountCredentialUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateApplicationLocalDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceApplicationLocalDomainAccountCredentialRead(ctx, d, m)
}

func resourceApplicationLocalDomainAccountCredentialDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteApplicationLocalDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceApplicationLocalDomainAccountCredentialImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 4 {
		return nil, errors.New("id must be <application_id>/<domain_id>/<account_id>/<type>")
	}
	id, ex, err := searchResourceApplicationLocalDomainAccountCredential(
		ctx, idSplit[0], idSplit[1], idSplit[2], idSplit[3], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find credential with id %s "+
			"(id must be <application_id>/<domain_id>/<account_id>/<type>)", d.Id())
	}
	cfg, err := readApplicationLocalDomainAccountCredentialOptions(ctx, idSplit[0], idSplit[1], idSplit[2], id, m)
	if err != nil {
		return nil, err
	}
	fillApplicationLocalDomainAccountCredential(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("application_id", idSplit[0]); tfErr != nil {
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

func searchResourceApplicationLocalDomainAccountCredential(
	ctx context.Context, applicationID, domainID, accountID, typeCred string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/applications/"+applicationID+"/localdomains/"+domainID+"/accounts/"+accountID+
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

func addApplicationLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainAccountCredentialJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateApplicationLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainAccountCredentialJSON(d)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteApplicationLocalDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Get("account_id").(string)+"/credentials/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareApplicationLocalDomainAccountCredentialJSON(
	d *schema.ResourceData,
) jsonCredential {
	// Application account credentials only support the "password" type
	// (no ssh_key/private_key/public_key, unlike device/domain account credentials).
	jsonData := jsonCredential{
		Type:     d.Get("type").(string),
		Password: d.Get("password").(string),
	}

	return jsonData
}

func readApplicationLocalDomainAccountCredentialOptions(
	ctx context.Context, applicationID, localDomainID, accountID, credentialID string, m interface{},
) (
	jsonCredential, error,
) {
	c := m.(*Client)
	var result jsonCredential
	body, code, err := c.newRequest(ctx,
		"/applications/"+applicationID+"/localdomains/"+localDomainID+
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
	credsID, found, err := searchResourceApplicationLocalDomainAccountCredential(
		ctx, applicationID, localDomainID, accountID, result.Type, m)
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

func fillApplicationLocalDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
}
