package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
			State: resourceExternalAuthLdapImport,
		},
		Schema: map[string]*schema.Schema{
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cn_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
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
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"timeout": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"ca_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
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
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"use_primary_auth_domain": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}
func resourveExternalAuthLdapVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_ldap not validate with api version %s", version)
}

func resourceExternalAuthLdapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthLdap(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get("authentication_name").(string)))
	}
	if !d.Get("is_anonymous_access").(bool) && (d.Get("login").(string) == "" || d.Get("password").(string) == "") {
		return diag.FromErr(fmt.Errorf("missing 'login' and/or 'password' on "+
			"externalauth_ldap %s", d.Get("authentication_name").(string)))
	}
	err = addExternalAuthLdap(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthLdap(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s can't find after POST", d.Get("authentication_name").(string)))
	}
	d.SetId(id)

	return resourceExternalAuthLdapRead(ctx, d, m)
}
func resourceExternalAuthLdapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
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
func resourceExternalAuthLdapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if !d.Get("is_anonymous_access").(bool) && (d.Get("login").(string) == "" || d.Get("password").(string) == "") {
		return diag.FromErr(fmt.Errorf("missing 'login' and/or 'password' on "+
			"externalauth_ldap %s", d.Get("authentication_name").(string)))
	}
	if err := updateExternalAuthLdap(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthLdapRead(ctx, d, m)
}
func resourceExternalAuthLdapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthLdap(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceExternalAuthLdapImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthLdap(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>", d.Id())
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
	ctx context.Context, authenticationName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonExternalAuthLdap
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.AuthenticationName == authenticationName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addExternalAuthLdap(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthLdapJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateExternalAuthLdap(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthLdapJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteExternalAuthLdap(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
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
		UsePrimaryAuthDomain: d.Get("use_primary_auth_domain").(bool),
		Timeout:              d.Get("timeout").(float64),
		AuthenticationName:   d.Get("authentication_name").(string),
		CACertificate:        d.Get("ca_certificate").(string),
		Certificate:          d.Get("certificate").(string),
		CNAttribute:          d.Get("cn_attribute").(string),
		Description:          d.Get("description").(string),
		LDAPBase:             d.Get("ldap_base").(string),
		Login:                d.Get("login").(string),
		LoginAttribute:       d.Get("login_attribute").(string),
		Host:                 d.Get("host").(string),
		Password:             d.Get("password").(string),
		Port:                 d.Get("port").(int),
		PrivateKey:           d.Get("private_key").(string),
		Type:                 "LDAP",
	}
}

func readExternalAuthLdapOptions(
	ctx context.Context, authenticationID string, m interface{}) (jsonExternalAuthLdap, error) {
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
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillExternalAuthLdap(d *schema.ResourceData, jsonData jsonExternalAuthLdap) {
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("cn_attribute", jsonData.CNAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
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
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("timeout", jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_certificate", jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate", jsonData.Certificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
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
	if tfErr := d.Set("private_key", jsonData.PrivateKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
