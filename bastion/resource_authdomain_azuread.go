package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonAuthDomainAzureAD struct {
	IsDefault          bool     `json:"is_default"`
	ID                 string   `json:"id,omitempty"`
	AuthDomainName     string   `json:"auth_domain_name"`
	Certificate        string   `json:"certificate"`
	ClientID           string   `json:"client_id"`
	ClientSecret       string   `json:"client_secret"`
	EntityID           string   `json:"entity_id"`
	Label              string   `json:"label"`
	DefaultEmailDomain string   `json:"default_email_domain"`
	DefaultLanguage    string   `json:"default_language"`
	Description        string   `json:"description"`
	DomainName         string   `json:"domain_name"`
	Passphrase         string   `json:"passphrase"`
	PrivateKey         string   `json:"private_key"`
	Type               string   `json:"type"`
	ExternalAuths      []string `json:"external_auths"`
	SecondaryAuth      []string `json:"secondary_auth"`
}

func resourceAuthDomainAzureAD() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthDomainAzureADCreate,
		ReadContext:   resourceAuthDomainAzureADRead,
		UpdateContext: resourceAuthDomainAzureADUpdate,
		DeleteContext: resourceAuthDomainAzureADDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthDomainAzureADImport,
		},
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auth_domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_email_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"de", "en", "es", "fr", "ru"}, false),
			},
			"entity_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"external_auths": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"label": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"passphrase": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{"private_key"},
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAuthDomainAzureADVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_authdomain_azuread not available with api version %s", version)
}

func resourceAuthDomainAzureADCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAuthDomainAzureAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get("domain_name").(string)))
	}
	err = addAuthDomainAzureAD(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainAzureAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s not found after POST", d.Get("domain_name").(string)))
	}
	d.SetId(id)

	return resourceAuthDomainAzureADRead(ctx, d, m)
}

func resourceAuthDomainAzureADRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAuthDomainAzureADOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAuthDomainAzureAD(d, cfg)
	}

	return nil
}

func resourceAuthDomainAzureADUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAuthDomainAzureAD(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAuthDomainAzureADRead(ctx, d, m)
}

func resourceAuthDomainAzureADDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAuthDomainAzureAD(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAuthDomainAzureADImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAuthDomainAzureAD(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>)", d.Id())
	}
	cfg, err := readAuthDomainAzureADOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillAuthDomainAzureAD(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAuthDomainAzureAD(
	ctx context.Context, domainName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authdomains/?q=domain_name"+domainName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAuthDomainAzureAD
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAuthDomainAzureAD(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainAzureADJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateAuthDomainAzureAD(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainAzureADJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAuthDomainAzureAD(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authdomains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareAuthDomainAzureADJSON(d *schema.ResourceData) jsonAuthDomainAzureAD {
	jsonData := jsonAuthDomainAzureAD{
		Type:               "AzureAD",
		DomainName:         d.Get("domain_name").(string),
		AuthDomainName:     d.Get("auth_domain_name").(string),
		ClientID:           d.Get("client_id").(string),
		DefaultEmailDomain: d.Get("default_email_domain").(string),
		DefaultLanguage:    d.Get("default_language").(string),
		EntityID:           d.Get("entity_id").(string),
		Label:              d.Get("label").(string),
		Certificate:        d.Get("certificate").(string),
		ClientSecret:       d.Get("client_secret").(string),
		Description:        d.Get("description").(string),
		IsDefault:          d.Get("is_default").(bool),
		Passphrase:         d.Get("passphrase").(string),
		PrivateKey:         d.Get("private_key").(string),
	}

	listExternalAuths := d.Get("external_auths").([]interface{})
	jsonData.ExternalAuths = make([]string, len(listExternalAuths))
	for i, v := range listExternalAuths {
		jsonData.ExternalAuths[i] = v.(string)
	}

	listSecondaryAuth := d.Get("secondary_auth").([]interface{})
	jsonData.SecondaryAuth = make([]string, len(listSecondaryAuth))
	for i, v := range listSecondaryAuth {
		jsonData.SecondaryAuth[i] = v.(string)
	}

	return jsonData
}

func readAuthDomainAzureADOptions(
	ctx context.Context, domainID string, m interface{},
) (
	jsonAuthDomainAzureAD, error,
) {
	c := m.(*Client)
	var result jsonAuthDomainAzureAD
	body, code, err := c.newRequest(ctx, "/authdomains/"+domainID, http.MethodGet, nil)
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

func fillAuthDomainAzureAD(d *schema.ResourceData, jsonData jsonAuthDomainAzureAD) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("client_id", jsonData.ClientID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("entity_id", jsonData.EntityID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_auths", jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("label", jsonData.Label); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_default", jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	// private_key hidden on API
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
}
