package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonAuthDomainSAML struct {
	ID                 string   `json:"id,omitempty"`
	DomainName         string   `json:"domain_name"`
	Type               string   `json:"type"`
	Description        string   `json:"description"`
	IsDefault          bool     `json:"is_default"`
	AuthDomainName     string   `json:"auth_domain_name"`
	ExternalAuths      []string `json:"external_auths"`
	SecondaryAuth      []string `json:"secondary_auth"`
	DefaultLanguage    string   `json:"default_language"`
	DefaultEmailDomain string   `json:"default_email_domain"`
	Label              string   `json:"label"`
	ForceAuthn         bool     `json:"force_authn"`

	IdpInitiatedURL string `json:"idp_initiated_url,omitempty"`
}

func resourceAuthDomainSAML() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthDomainSAMLCreate,
		ReadContext:   resourceAuthDomainSAMLRead,
		UpdateContext: resourceAuthDomainSAMLUpdate,
		DeleteContext: resourceAuthDomainSAMLDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthDomainSAMLImport,
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
			"default_email_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"de", "en", "es", "fr", "ru"}, false),
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"force_authn": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"idp_initiated_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAuthDomainSAMLVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_authdomain_saml not available with api version %s", version)
}

func resourceAuthDomainSAMLCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAuthDomainSAML(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get("domain_name").(string)))
	}
	err = addAuthDomainSAML(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainSAML(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s not found after POST", d.Get("domain_name").(string)))
	}
	d.SetId(id)

	return resourceAuthDomainSAMLRead(ctx, d, m)
}

func resourceAuthDomainSAMLRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAuthDomainSAMLOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAuthDomainSAML(d, cfg)
	}

	return nil
}

func resourceAuthDomainSAMLUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAuthDomainSAML(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAuthDomainSAMLRead(ctx, d, m)
}

func resourceAuthDomainSAMLDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAuthDomainSAML(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAuthDomainSAMLImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAuthDomainSAML(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>)", d.Id())
	}
	cfg, err := readAuthDomainSAMLOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillAuthDomainSAML(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAuthDomainSAML(
	ctx context.Context, domainName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authdomains/?q=domain_name="+domainName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAuthDomainSAML
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAuthDomainSAML(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainSAMLJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateAuthDomainSAML(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainSAMLJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAuthDomainSAML(
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

func prepareAuthDomainSAMLJSON(d *schema.ResourceData) jsonAuthDomainSAML {
	jsonData := jsonAuthDomainSAML{
		DomainName:         d.Get("domain_name").(string),
		Type:               "SAML",
		Description:        d.Get("description").(string),
		IsDefault:          d.Get("is_default").(bool),
		AuthDomainName:     d.Get("auth_domain_name").(string),
		DefaultLanguage:    d.Get("default_language").(string),
		DefaultEmailDomain: d.Get("default_email_domain").(string),
		Label:              d.Get("label").(string),
		ForceAuthn:         d.Get("force_authn").(bool),
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

func readAuthDomainSAMLOptions(
	ctx context.Context, domainID string, m interface{},
) (
	jsonAuthDomainSAML, error,
) {
	c := m.(*Client)
	var result jsonAuthDomainSAML
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

func fillAuthDomainSAML(d *schema.ResourceData, jsonData jsonAuthDomainSAML) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
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
	if tfErr := d.Set("force_authn", jsonData.ForceAuthn); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_default", jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_initiated_url", jsonData.IdpInitiatedURL); tfErr != nil {
		panic(tfErr)
	}
}
