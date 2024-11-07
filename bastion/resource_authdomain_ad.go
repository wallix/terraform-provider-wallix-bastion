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

type jsonAuthDomainAD struct {
	CheckX509SanEmail    bool     `json:"check_x509_san_email"`
	IsDefault            bool     `json:"is_default"`
	ID                   string   `json:"id,omitempty"`
	AuthDomainName       string   `json:"auth_domain_name"`
	DefaultEmailDomain   string   `json:"default_email_domain"`
	DefaultLanguage      string   `json:"default_language"`
	Description          string   `json:"description"`
	DisplayNameAttribute string   `json:"display_name_attribute"`
	DomainName           string   `json:"domain_name"`
	EmailAttribute       string   `json:"email_attribute"`
	GroupAttribute       string   `json:"group_attribute"`
	LanguageAttribute    string   `json:"language_attribute"`
	PubKeyAttribute      string   `json:"pubkey_attribute"`
	SanDomainName        string   `json:"san_domain_name"`
	Type                 string   `json:"type"`
	X509Condition        string   `json:"x509_condition"`
	X509SearchFilter     string   `json:"x509_search_filter"`
	ExternalAuths        []string `json:"external_auths"`
	SecondaryAuth        []string `json:"secondary_auth"`
}

func resourceAuthDomainAD() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthDomainADCreate,
		ReadContext:   resourceAuthDomainADRead,
		UpdateContext: resourceAuthDomainADUpdate,
		DeleteContext: resourceAuthDomainADDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthDomainADImport,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"check_x509_san_email": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"display_name_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"language_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pubkey_attribute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"san_domain_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"x509_condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"x509_search_filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAuthDomainADVersionCheck(version string) error {
	if bchk.InSlice(version, versions38Plus()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_authdomain_ad not available with api version %s", version)
}

func resourceAuthDomainADCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAuthDomainAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get("domain_name").(string)))
	}
	err = addAuthDomainAD(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s not found after POST", d.Get("domain_name").(string)))
	}
	d.SetId(id)

	return resourceAuthDomainADRead(ctx, d, m)
}

func resourceAuthDomainADRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAuthDomainADOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAuthDomainAD(d, cfg)
	}

	return nil
}

func resourceAuthDomainADUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAuthDomainADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAuthDomainAD(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAuthDomainADRead(ctx, d, m)
}

func resourceAuthDomainADDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAuthDomainAD(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAuthDomainADImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceAuthDomainADVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAuthDomainAD(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>)", d.Id())
	}
	cfg, err := readAuthDomainADOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillAuthDomainAD(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAuthDomainAD(
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
	var results []jsonAuthDomainAD
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAuthDomainAD(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainADJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateAuthDomainAD(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainADJSON(d)
	body, code, err := c.newRequest(ctx, "/authdomains/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAuthDomainAD(
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

func prepareAuthDomainADJSON(d *schema.ResourceData) jsonAuthDomainAD {
	jsonData := jsonAuthDomainAD{
		Type:                 "AD",
		DomainName:           d.Get("domain_name").(string),
		AuthDomainName:       d.Get("auth_domain_name").(string),
		DefaultEmailDomain:   d.Get("default_email_domain").(string),
		DefaultLanguage:      d.Get("default_language").(string),
		Description:          d.Get("description").(string),
		CheckX509SanEmail:    d.Get("check_x509_san_email").(bool),
		DisplayNameAttribute: d.Get("display_name_attribute").(string),
		EmailAttribute:       d.Get("email_attribute").(string),
		GroupAttribute:       d.Get("group_attribute").(string),
		IsDefault:            d.Get("is_default").(bool),
		LanguageAttribute:    d.Get("language_attribute").(string),
		PubKeyAttribute:      d.Get("pubkey_attribute").(string),
		SanDomainName:        d.Get("san_domain_name").(string),
		X509Condition:        d.Get("x509_condition").(string),
		X509SearchFilter:     d.Get("x509_search_filter").(string),
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

func readAuthDomainADOptions(
	ctx context.Context, domainID string, m interface{},
) (
	jsonAuthDomainAD, error,
) {
	c := m.(*Client)
	var result jsonAuthDomainAD
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

func fillAuthDomainAD(d *schema.ResourceData, jsonData jsonAuthDomainAD) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_auths", jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("check_x509_san_email", jsonData.CheckX509SanEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("group_attribute", jsonData.GroupAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("display_name_attribute", jsonData.DisplayNameAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("email_attribute", jsonData.EmailAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_default", jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("language_attribute", jsonData.LanguageAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("pubkey_attribute", jsonData.PubKeyAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("san_domain_name", jsonData.SanDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("x509_condition", jsonData.X509Condition); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("x509_search_filter", jsonData.X509SearchFilter); tfErr != nil {
		panic(tfErr)
	}
}
