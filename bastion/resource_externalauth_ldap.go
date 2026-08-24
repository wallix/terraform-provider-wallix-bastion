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

type jsonExternalAuthLdap struct {
	IsActiveDirectory    bool    `json:"is_active_directory"`
	IsAnonymousAccess    bool    `json:"is_anonymous_access"`
	IsProtectedUser      bool    `json:"is_protected_user"`
	IsSSL                bool    `json:"is_ssl"`
	IsStartTLS           bool    `json:"is_starttls"`
	UsePrimaryAuthDomain bool    `json:"use_primary_auth_domain"`
	Port                 int     `json:"port"`
	Timeout              float64 `json:"timeout"`
	ID                   string  `json:"id,omitempty"`
	AuthenticationName   string  `json:"authentication_name"`
	CACertificate        string  `json:"ca_certificate"`
	Certificate          string  `json:"certificate"`
	CNAttribute          string  `json:"cn_attribute"`
	Description          string  `json:"description"`
	LDAPBase             string  `json:"ldap_base"`
	Login                string  `json:"login,omitempty"`
	LoginAttribute       string  `json:"login_attribute"`
	Host                 string  `json:"host"`
	Passphrase           string  `json:"passphrase,omitempty"`
	Password             string  `json:"password,omitempty"`
	PrivateKey           string  `json:"private_key"`
	Type                 string  `json:"type"`
}

func resourceExternalAuthLdap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalAuthLdapCreate,
		ReadContext:   resourceExternalAuthLdapRead,
		UpdateContext: resourceExternalAuthLdapUpdate,
		DeleteContext: resourceExternalAuthLdapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceExternalAuthLdapImport,
		},
		Schema: map[string]*schema.Schema{
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"cn_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Required: true,
			},
			"ldap_base": {
				Type:     schema.TypeString,
				Required: true,
			},
			"login_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			skPort: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			skTimeout: {
				Type:     schema.TypeFloat,
				Required: true,
			},
			skCACertificate: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skCertificate: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_active_directory": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_anonymous_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_protected_user": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_ssl": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_starttls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"login": {
				Type:     schema.TypeString,
				Optional: true,
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
			},
			skUsePrimaryAuthDomain: {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceExternalAuthLdapVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_ldap not available with api version %s", version)
}

func resourceExternalAuthLdapCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthLdap(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get(skAuthenticationName).(string)))
	}
	if !d.Get("is_anonymous_access").(bool) && (d.Get("login").(string) == "" || d.Get(skPassword).(string) == "") {
		return diag.FromErr(fmt.Errorf("missing 'login' and/or 'password' on "+
			"externalauth_ldap %s", d.Get(skAuthenticationName).(string)))
	}
	id, err := addExternalAuthLdap(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceExternalAuthLdap(ctx, d.Get(skAuthenticationName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("authentication_name %s not found after POST", d.Get(skAuthenticationName).(string)))
		}
	}
	d.SetId(id)

	return resourceExternalAuthLdapRead(ctx, d, m)
}

func resourceExternalAuthLdapRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readExternalAuthLdapOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillExternalAuthLdap(d, cfg)
	}

	return nil
}

func resourceExternalAuthLdapUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if !d.Get("is_anonymous_access").(bool) && (d.Get("login").(string) == "" || d.Get(skPassword).(string) == "") {
		return diag.FromErr(fmt.Errorf("missing 'login' and/or 'password' on "+
			"externalauth_ldap %s", d.Get(skAuthenticationName).(string)))
	}
	if err := updateExternalAuthLdap(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthLdapRead(ctx, d, m)
}

func resourceExternalAuthLdapDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthLdap(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceExternalAuthLdapImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthLdap(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>)", d.Id())
	}
	cfg, err := readExternalAuthLdapOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillExternalAuthLdap(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceExternalAuthLdap(
	ctx context.Context, authenticationName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/?q=authentication_name="+authenticationName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonExternalAuthLdap
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addExternalAuthLdap(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareExternalAuthLdapJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateExternalAuthLdap(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthLdapJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteExternalAuthLdap(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareExternalAuthLdapJSON(d *schema.ResourceData) jsonExternalAuthLdap {
	return jsonExternalAuthLdap{
		IsActiveDirectory:    d.Get("is_active_directory").(bool),
		IsAnonymousAccess:    d.Get("is_anonymous_access").(bool),
		IsProtectedUser:      d.Get("is_protected_user").(bool),
		IsSSL:                d.Get("is_ssl").(bool),
		IsStartTLS:           d.Get("is_starttls").(bool),
		UsePrimaryAuthDomain: d.Get(skUsePrimaryAuthDomain).(bool),
		Timeout:              d.Get(skTimeout).(float64),
		AuthenticationName:   d.Get(skAuthenticationName).(string),
		CACertificate:        d.Get(skCACertificate).(string),
		Certificate:          d.Get(skCertificate).(string),
		CNAttribute:          d.Get("cn_attribute").(string),
		Description:          d.Get(skDescription).(string),
		LDAPBase:             d.Get("ldap_base").(string),
		Login:                d.Get("login").(string),
		LoginAttribute:       d.Get("login_attribute").(string),
		Host:                 d.Get(skHost).(string),
		Password:             d.Get(skPassword).(string),
		Passphrase:           d.Get(skPassphrase).(string),
		Port:                 d.Get(skPort).(int),
		PrivateKey:           d.Get(skPrivateKey).(string),
		Type:                 "LDAP",
	}
}

func readExternalAuthLdapOptions(
	ctx context.Context, authenticationID string, m interface{},
) (
	jsonExternalAuthLdap, error,
) {
	c := m.(*Client)
	var result jsonExternalAuthLdap
	body, code, err := c.newRequest(ctx, "/externalauths/"+authenticationID, http.MethodGet, nil)
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

func fillExternalAuthLdap(d *schema.ResourceData, jsonData jsonExternalAuthLdap) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("cn_attribute", jsonData.CNAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ldap_base", jsonData.LDAPBase); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("login", jsonData.Login); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("login_attribute", jsonData.LoginAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skTimeout, jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCACertificate, jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_active_directory", jsonData.IsActiveDirectory); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_anonymous_access", jsonData.IsAnonymousAccess); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_protected_user", jsonData.IsProtectedUser); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_ssl", jsonData.IsSSL); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_starttls", jsonData.IsStartTLS); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
