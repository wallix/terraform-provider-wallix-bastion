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

func resourceDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainAccountCredentialCreate,
		ReadContext:   resourceDomainAccountCredentialRead,
		UpdateContext: resourceDomainAccountCredentialUpdate,
		DeleteContext: resourceDomainAccountCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainAccountCredentialImport,
		},
		Schema: map[string]*schema.Schema{
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
			"propagate_credential_change": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain_account_credential not available with api version %s", version)
}

func resourceDomainAccountCredentialCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDomain, err := readDomainOptions(ctx, d.Get(skDomainID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s doesn't exists", d.Get(skDomainID).(string)))
	}
	cfgAccount, err := readDomainAccountOptions(ctx, d.Get(skDomainID).(string), d.Get(skAccountID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgAccount.ID == "" {
		return diag.FromErr(fmt.Errorf("account_id with ID %s on domain_id %s doesn't exists",
			d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
	}
	_, ex, err := searchResourceDomainAccountCredential(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s already exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
	}
	id, err := addDomainAccountCredential(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDomainAccountCredential(ctx,
			d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"credential type %s on account_id %s, domain_id %s not found after POST",
				d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
		}
	}
	d.SetId(id)

	return resourceDomainAccountCredentialRead(ctx, d, m)
}

func resourceDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDomainAccountCredentialOptions(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDomainAccountCredential(d, cfg)
	}

	return nil
}

func resourceDomainAccountCredentialUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDomainAccountCredentialRead(ctx, d, m)
}

func resourceDomainAccountCredentialDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainAccountCredentialImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return nil, errors.New("id must be <domain_id>/<account_id>/<type>")
	}
	id, ex, err := searchResourceDomainAccountCredential(ctx, idSplit[0], idSplit[1], idSplit[2], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find credential with id %s "+
			"(id must be <domain_id>/<account_id>/<type>)", d.Id())
	}
	cfg, err := readDomainAccountCredentialOptions(ctx, idSplit[0], idSplit[1], id, m)
	if err != nil {
		return nil, err
	}
	fillDomainAccountCredential(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set(skDomainID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountID, idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDomainAccountCredential(
	ctx context.Context, domainID, accountID, typeCred string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/"+accountID+
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

func addDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	propagate := d.Get("propagate_credential_change").(bool)
	jsonData := prepareDomainAccountCredentialJSON(d, propagate, true)

	body, headers, code, err := c.newRequestWithHeaders(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Get(skAccountID).(string)+"/credentials/",
		http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	id := headers.Get("X-Object-Id")

	if propagate {
		accountID := d.Get(skAccountID)
		jsonDataPropagate := prepareDomainAccountCredentialJSON(d, propagate, false)

		body, code, err = c.newRequest(ctx,
			fmt.Sprintf("/accountchangepassword/%s/password", accountID),
			http.MethodPut, jsonDataPropagate)
		if err != nil {
			return "", err
		}
		if code != http.StatusOK && code != http.StatusNoContent {
			return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
		}
	}

	return id, nil
}

func updateDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	// Extract the client from the meta parameter
	client, ok := m.(*Client)
	if !ok {
		return errors.New("failed to cast interface to *Client")
	}

	// Get the account ID and domain ID from the resource data
	accountID, ok := d.Get(skAccountID).(string)
	if !ok {
		return errors.New("failed to get account_id from resource data")
	}

	domainID, ok := d.Get(skDomainID).(string)
	if !ok {
		return errors.New("failed to get domain_id from resource data")
	}

	// Check the value of propagate_credential_change option
	propagate := d.Get("propagate_credential_change").(bool)

	// Prepare the JSON data for the request
	jsonData := prepareDomainAccountCredentialJSON(d, propagate, false)
	var url string
	if propagate {
		url = fmt.Sprintf("/accountchangepassword/%s/password", accountID)
	} else {
		url = fmt.Sprintf("/domains/%s/accounts/%s/credentials/%s", domainID, accountID, d.Id())
	}

	// Make the appropriate request based on the propagate_credential_change value
	body, code, err := client.newRequest(ctx, url, http.MethodPut, jsonData)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API didn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Get(skAccountID).(string)+"/credentials/"+d.Id(),
		http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDomainAccountCredentialJSON(
	d *schema.ResourceData,
	propagate bool,
	propagateAdd bool,
) jsonCredential {
	var jsonData jsonCredential

	if propagate {
		// Only include the password key
		jsonData.Password = d.Get(skPassword).(string)
		if propagateAdd {
			jsonData.Type = d.Get(skType).(string)
		}
	} else {
		// Include the type and other fields based on the type
		jsonData.Type = d.Get(skType).(string)

		switch jsonData.Type {
		case skPassword:
			jsonData.Password = d.Get(skPassword).(string)
		case "ssh_key":
			jsonData.PrivateKey = d.Get(skPrivateKey).(string)
			jsonData.Passphrase = d.Get(skPassphrase).(string)
		}
	}

	return jsonData
}

func readDomainAccountCredentialOptions(
	ctx context.Context, domainID, accountID, credentialID string, m interface{},
) (
	jsonCredential, error,
) {
	c := m.(*Client)
	var result jsonCredential
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/"+accountID+"/credentials/"+credentialID,
		http.MethodGet, nil)
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
	credsID, found, err := searchResourceDomainAccountCredential(ctx, domainID, accountID, result.Type, m)
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

func fillDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPublicKey, jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
