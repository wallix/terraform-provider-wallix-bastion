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

type jsonExternalAuthTacacs struct {
	Port                 int    `json:"port"`
	ID                   string `json:"id,omitempty"`
	AuthenticationName   string `json:"authentication_name"`
	Description          string `json:"description"`
	Host                 string `json:"host"`
	Secret               string `json:"secret"`
	Type                 string `json:"type"`
	UsePrimaryAuthDomain bool   `json:"use_primary_auth_domain"`
}

func resourceExternalAuthTacacs() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalAuthTacacsCreate,
		ReadContext:   resourceExternalAuthTacacsRead,
		UpdateContext: resourceExternalAuthTacacsUpdate,
		DeleteContext: resourceExternalAuthTacacsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceExternalAuthTacacsImport,
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
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"description": {
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
func resourceExternalAuthTacacsVersionCheck(version string) error {
	if bchk.StringInSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_tacacs not validate with api version %s", version)
}

func resourceExternalAuthTacacsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthTacacs(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get("authentication_name").(string)))
	}
	err = addExternalAuthTacacs(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthTacacs(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s can't find after POST", d.Get("authentication_name").(string)))
	}
	d.SetId(id)

	return resourceExternalAuthTacacsRead(ctx, d, m)
}
func resourceExternalAuthTacacsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readExternalAuthTacacsOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillExternalAuthTacacs(d, cfg)
	}

	return nil
}
func resourceExternalAuthTacacsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateExternalAuthTacacs(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthTacacsRead(ctx, d, m)
}
func resourceExternalAuthTacacsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthTacacs(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceExternalAuthTacacsImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthTacacs(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>", d.Id())
	}
	cfg, err := readExternalAuthTacacsOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillExternalAuthTacacs(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceExternalAuthTacacs(
	ctx context.Context, authenticationName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/externalauths/?fields=authentication_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonExternalAuthTacacs
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

func addExternalAuthTacacs(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthTacacsJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateExternalAuthTacacs(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthTacacsJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteExternalAuthTacacs(ctx context.Context, d *schema.ResourceData, m interface{}) error {
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

func prepareExternalAuthTacacsJSON(d *schema.ResourceData) jsonExternalAuthTacacs {
	return jsonExternalAuthTacacs{
		AuthenticationName:   d.Get("authentication_name").(string),
		Host:                 d.Get("host").(string),
		Port:                 d.Get("port").(int),
		Secret:               d.Get("secret").(string),
		Description:          d.Get("description").(string),
		UsePrimaryAuthDomain: d.Get("use_primary_auth_domain").(bool),
		Type:                 "TACACS+",
	}
}

func readExternalAuthTacacsOptions(
	ctx context.Context, authenticationID string, m interface{}) (jsonExternalAuthTacacs, error) {
	c := m.(*Client)
	var result jsonExternalAuthTacacs
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

func fillExternalAuthTacacs(d *schema.ResourceData, jsonData jsonExternalAuthTacacs) {
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
