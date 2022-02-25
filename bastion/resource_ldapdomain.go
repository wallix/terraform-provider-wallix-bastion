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

type jsonLdapDomain struct {
	CheckX509SanEmail    bool     `json:"check_x509_san_email"`
	IsDefault            bool     `json:"is_default"`
	DomainName           string   `json:"domain_name,omitempty"`
	DefaultLanguage      string   `json:"default_language"`
	DefaultEmailDomain   string   `json:"default_email_domain"`
	Description          string   `json:"description"`
	DisplayNameAttribute string   `json:"display_name_attribute"`
	EmailAttribute       string   `json:"email_attribute"`
	LdapDomainName       string   `json:"ldap_domain_name"`
	LanguageAttribute    string   `json:"language_attribute"`
	GroupAttribute       string   `json:"group_attribute"`
	SanDomainName        string   `json:"san_domain_name"`
	X509Condition        string   `json:"x509_condition"`
	X509SearchFilter     string   `json:"x509_search_filter"`
	ExternalLdaps        []string `json:"external_ldaps"`
	SecondaryAuth        []string `json:"secondary_auth"`
}

func resourceLdapDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLdapDomainCreate,
		ReadContext:   resourceLdapDomainRead,
		UpdateContext: resourceLdapDomainUpdate,
		DeleteContext: resourceLdapDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceLdapDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ldap_domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"external_ldaps": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"de", "en", "es", "fr", "ru"}, false),
			},
			"default_email_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Optional: true,
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
			"san_domain_name": {
				Type:     schema.TypeString,
				Optional: true,
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
func resourceLdapDomainVersionCheck(version string) error {
	if bchk.StringInSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_ldapdomain not validate with api version %s", version)
}

func resourceLdapDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	ex, err := checkResourceLdapDomainExists(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("domain_name %s already exists", d.Get("domain_name").(string)))
	}
	err = addLdapDomain(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("domain_name").(string))

	return resourceLdapDomainRead(ctx, d, m)
}
func resourceLdapDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readLdapDomainOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.DomainName == "" {
		d.SetId("")
	} else {
		fillLdapDomain(d, cfg)
	}

	return nil
}
func resourceLdapDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceLdapDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateLdapDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceLdapDomainRead(ctx, d, m)
}
func resourceLdapDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteLdapDomain(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceLdapDomainImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceLdapDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	ex, err := checkResourceLdapDomainExists(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find domain_name with id %s (id must be <domain_name>", d.Id())
	}
	cfg, err := readLdapDomainOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillLdapDomain(d, cfg)
	result := make([]*schema.ResourceData, 1)
	result[0] = d

	return result, nil
}

func checkResourceLdapDomainExists(ctx context.Context, domainName string, m interface{}) (bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/ldapdomains/"+domainName, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}

	return true, nil
}

func addLdapDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareLdapDomainJSON(d, true)
	body, code, err := c.newRequest(ctx, "/ldapdomains/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateLdapDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareLdapDomainJSON(d, false)
	body, code, err := c.newRequest(ctx, "/ldapdomains/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteLdapDomain(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/ldapdomains/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareLdapDomainJSON(d *schema.ResourceData, newResource bool) jsonLdapDomain {
	jsonData := jsonLdapDomain{
		LdapDomainName:       d.Get("ldap_domain_name").(string),
		DefaultLanguage:      d.Get("default_language").(string),
		DefaultEmailDomain:   d.Get("default_email_domain").(string),
		Description:          d.Get("description").(string),
		CheckX509SanEmail:    d.Get("check_x509_san_email").(bool),
		DisplayNameAttribute: d.Get("display_name_attribute").(string),
		EmailAttribute:       d.Get("email_attribute").(string),
		GroupAttribute:       d.Get("group_attribute").(string),
		IsDefault:            d.Get("is_default").(bool),
		LanguageAttribute:    d.Get("language_attribute").(string),
		SanDomainName:        d.Get("san_domain_name").(string),
		X509Condition:        d.Get("x509_condition").(string),
		X509SearchFilter:     d.Get("x509_search_filter").(string),
	}
	if newResource {
		jsonData.DomainName = d.Get("domain_name").(string)
	}
	for _, v := range d.Get("external_ldaps").([]interface{}) {
		jsonData.ExternalLdaps = append(jsonData.ExternalLdaps, v.(string))
	}
	jsonData.SecondaryAuth = make([]string, 0)
	for _, v := range d.Get("secondary_auth").([]interface{}) {
		jsonData.SecondaryAuth = append(jsonData.SecondaryAuth, v.(string))
	}

	return jsonData
}

func readLdapDomainOptions(
	ctx context.Context, domainName string, m interface{}) (jsonLdapDomain, error) {
	c := m.(*Client)
	var result jsonLdapDomain
	body, code, err := c.newRequest(ctx, "/ldapdomains/"+domainName, http.MethodGet, nil)
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

func fillLdapDomain(d *schema.ResourceData, jsonData jsonLdapDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ldap_domain_name", jsonData.LdapDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_ldaps", jsonData.ExternalLdaps); tfErr != nil {
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
