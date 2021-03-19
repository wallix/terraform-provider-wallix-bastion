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

type jsonExternalAuthKerberos struct {
	UsePrimaryAuthDomain bool   `json:"use_primary_auth_domain"`
	Port                 int    `json:"port"`
	ID                   string `json:"id,omitempty"`
	AuthenticationName   string `json:"authentication_name"`
	Description          string `json:"description"`
	Host                 string `json:"host"`
	KerDomController     string `json:"ker_dom_controller"`
	LoginAttribute       string `json:"login_attribute"`
	Type                 string `json:"type"`
}

func resourceExternalAuthKerberos() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalAuthKerberosCreate,
		ReadContext:   resourceExternalAuthKerberosRead,
		UpdateContext: resourceExternalAuthKerberosUpdate,
		DeleteContext: resourceExternalAuthKerberosDelete,
		Importer: &schema.ResourceImporter{
			State: resourceExternalAuthKerberosImport,
		},
		Schema: map[string]*schema.Schema{
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ker_dom_controller": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"kerberos_password": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_attribute": {
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
func resourceExternalAuthKerberosVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_kerberos not validate with api version %s", version)
}

func resourceExternalAuthKerberosCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthKerberos(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get("authentication_name").(string)))
	}
	err = addExternalAuthKerberos(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthKerberos(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s can't find after POST", d.Get("authentication_name").(string)))
	}
	d.SetId(id)

	return resourceExternalAuthKerberosRead(ctx, d, m)
}
func resourceExternalAuthKerberosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readExternalAuthKerberosOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillExternalAuthKerberos(d, cfg)
	}

	return nil
}
func resourceExternalAuthKerberosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateExternalAuthKerberos(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthKerberosRead(ctx, d, m)
}
func resourceExternalAuthKerberosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthKerberos(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceExternalAuthKerberosImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceExternalAuthKerberosVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthKerberos(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>", d.Id())
	}
	cfg, err := readExternalAuthKerberosOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillExternalAuthKerberos(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceExternalAuthKerberos(
	ctx context.Context, authenticationName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/?fields=authentication_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonExternalAuthKerberos
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

func addExternalAuthKerberos(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthKerberosJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateExternalAuthKerberos(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthKerberosJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteExternalAuthKerberos(ctx context.Context, d *schema.ResourceData, m interface{}) error {
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

func prepareExternalAuthKerberosJSON(d *schema.ResourceData) jsonExternalAuthKerberos {
	jsonData := jsonExternalAuthKerberos{
		AuthenticationName:   d.Get("authentication_name").(string),
		Host:                 d.Get("host").(string),
		KerDomController:     d.Get("ker_dom_controller").(string),
		Port:                 d.Get("port").(int),
		Description:          d.Get("description").(string),
		LoginAttribute:       d.Get("login_attribute").(string),
		UsePrimaryAuthDomain: d.Get("use_primary_auth_domain").(bool),
		Type:                 "KERBEROS",
	}
	if d.Get("kerberos_password").(bool) {
		jsonData.Type = "KERBEROS-PASSWORD"
	}

	return jsonData
}

func readExternalAuthKerberosOptions(
	ctx context.Context, authenticationID string, m interface{}) (jsonExternalAuthKerberos, error) {
	c := m.(*Client)
	var result jsonExternalAuthKerberos
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

func fillExternalAuthKerberos(d *schema.ResourceData, jsonData jsonExternalAuthKerberos) {
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ker_dom_controller", jsonData.KerDomController); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("login_attribute", jsonData.LoginAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Type == "KERBEROS-PASSWORD" {
		if tfErr := d.Set("kerberos_password", true); tfErr != nil {
			panic(tfErr)
		}
	}
}
