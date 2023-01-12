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

type jsonExternalAuthSaml struct {
	Timeout                    float64 `json:"timeout"`
	ID                         string  `json:"id,omitempty"`
	AuthenticationName         string  `json:"authentication_name"`
	Certificate                string  `json:"certificate"`
	Description                string  `json:"description"`
	IDPEntityID                string  `json:"idp_entity_id,omitempty"`
	IDPMetadata                string  `json:"idp_metadata"`
	Passphrase                 string  `json:"passphrase,omitempty"`
	PrivateKey                 string  `json:"private_key"`
	SamlRequestMethod          string  `json:"saml_request_method,omitempty"`
	SamlRequestURL             string  `json:"saml_request_url,omitempty"`
	SPAssertionConsumerService string  `json:"sp_assertion_consumer_service,omitempty"`
	SPEntityID                 string  `json:"sp_entity_id,omitempty"`
	SPMetadata                 string  `json:"sp_metadata,omitempty"`
	SPSingleLogoutService      string  `json:"sp_single_logout_service,omitempty"`
	Type                       string  `json:"type"`
}

func resourceExternalAuthSaml() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalAuthSamlCreate,
		ReadContext:   resourceExternalAuthSamlRead,
		UpdateContext: resourceExternalAuthSamlUpdate,
		DeleteContext: resourceExternalAuthSamlDelete,
		Importer: &schema.ResourceImporter{
			State: resourceExternalAuthSamlImport,
		},
		Schema: map[string]*schema.Schema{
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"idp_metadata": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timeout": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(1, 900),
			},
			"certificate": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
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
			"idp_entity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_request_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_request_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_entity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_assertion_consumer_service": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sp_single_logout_service": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceExternalAuthSamlVersionCheck(version string) error {
	if bchk.InSlice(version, []string{VersionWallixAPI38}) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_saml not available with api version %s", version)
}

func resourceExternalAuthSamlCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthSaml(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get("authentication_name").(string)))
	}
	err = addExternalAuthSaml(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthSaml(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s not found after POST", d.Get("authentication_name").(string)))
	}
	d.SetId(id)

	return resourceExternalAuthSamlRead(ctx, d, m)
}

func resourceExternalAuthSamlRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readExternalAuthSamlOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillExternalAuthSaml(d, cfg)
	}

	return nil
}

func resourceExternalAuthSamlUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateExternalAuthSaml(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthSamlRead(ctx, d, m)
}

func resourceExternalAuthSamlDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthSaml(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceExternalAuthSamlImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceExternalAuthSamlVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthSaml(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>", d.Id())
	}
	cfg, err := readExternalAuthSamlOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillExternalAuthSaml(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceExternalAuthSaml(
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
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonExternalAuthSaml
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addExternalAuthSaml(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthSamlJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateExternalAuthSaml(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthSamlJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteExternalAuthSaml(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
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

func prepareExternalAuthSamlJSON(d *schema.ResourceData) jsonExternalAuthSaml {
	return jsonExternalAuthSaml{
		AuthenticationName: d.Get("authentication_name").(string),
		Type:               "SAML",
		IDPMetadata:        d.Get("idp_metadata").(string),
		Timeout:            d.Get("timeout").(float64),
		Certificate:        d.Get("certificate").(string),
		Description:        d.Get("description").(string),
		Passphrase:         d.Get("passphrase").(string),
		PrivateKey:         d.Get("private_key").(string),
	}
}

func readExternalAuthSamlOptions(
	ctx context.Context, authenticationID string, m interface{},
) (
	jsonExternalAuthSaml, error,
) {
	c := m.(*Client)
	var result jsonExternalAuthSaml
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

func fillExternalAuthSaml(d *schema.ResourceData, jsonData jsonExternalAuthSaml) {
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_metadata", jsonData.IDPMetadata); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("timeout", jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_entity_id", jsonData.IDPEntityID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("saml_request_url", jsonData.SamlRequestURL); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("saml_request_method", jsonData.SamlRequestMethod); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_metadata", jsonData.SPMetadata); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_entity_id", jsonData.SPEntityID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_assertion_consumer_service", jsonData.SPAssertionConsumerService); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sp_single_logout_service", jsonData.SPSingleLogoutService); tfErr != nil {
		panic(tfErr)
	}
}
