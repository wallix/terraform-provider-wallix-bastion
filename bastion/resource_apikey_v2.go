package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/mod/semver"
)

type jsonAPIKeyV2 struct {
	ID           string `json:"id,omitempty"`
	APIKeyName   string `json:"apikey_name"`
	Description  string `json:"description"`
	Profile      string `json:"profile,omitempty"`
	APIKey       string `json:"apikey,omitempty"`
	IPLimitation string `json:"ip_limitation,omitempty"`
}

func resourceAPIKeyV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyV2Create,
		ReadContext:   resourceAPIKeyV2Read,
		UpdateContext: resourceAPIKeyV2Update,
		DeleteContext: resourceAPIKeyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIKeyV2Import,
		},
		Schema: map[string]*schema.Schema{
			"apikey_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_limitation": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"apikey": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// apikeys-v2 doesn't exist before API v3.12.
func resourceAPIKeyV2VersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_apikey_v2 not available with api version %s", version)
}

func resourceAPIKeyV2Create(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAPIKeyV2(ctx, d.Get("apikey_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("apikey_name %s already exists", d.Get("apikey_name").(string)))
	}
	id, err := addAPIKeyV2(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceAPIKeyV2(ctx, d.Get("apikey_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("apikey_name %s not found after POST", d.Get("apikey_name").(string)))
		}
	}
	d.SetId(id)

	return resourceAPIKeyV2Read(ctx, d, m)
}

func resourceAPIKeyV2Read(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAPIKeyV2Options(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAPIKeyV2(d, cfg)
	}

	return nil
}

func resourceAPIKeyV2Update(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAPIKeyV2(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAPIKeyV2Read(ctx, d, m)
}

func resourceAPIKeyV2Delete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAPIKeyV2(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAPIKeyV2Import(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAPIKeyV2(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find apikey_name with id %s (id must be <apikey_name>)", d.Id())
	}
	cfg, err := readAPIKeyV2Options(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillAPIKeyV2(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAPIKeyV2(
	ctx context.Context, apikeyName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/apikeys-v2/?q=apikey_name="+apikeyName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAPIKeyV2
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAPIKeyV2(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareAPIKeyV2JSON(d, true)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/apikeys-v2/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateAPIKeyV2(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAPIKeyV2JSON(d, false)
	body, code, err := c.newRequest(ctx, "/apikeys-v2/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAPIKeyV2(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/apikeys-v2/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// prepareAPIKeyV2JSON builds the request body. On update, the API only accepts
// apikey_name and description ("profile" and "ip_limitation" are immutable, per
// apikey_v2_put in openapi.json and confirmed against a live Bastion), so they
// must be omitted or the API rejects the request as an unexpected property.
func prepareAPIKeyV2JSON(d *schema.ResourceData, newResource bool) jsonAPIKeyV2 {
	jsonData := jsonAPIKeyV2{
		APIKeyName:  d.Get("apikey_name").(string),
		Description: d.Get("description").(string),
	}

	if newResource {
		jsonData.Profile = d.Get("profile").(string)
		jsonData.IPLimitation = d.Get("ip_limitation").(string)
	}

	return jsonData
}

func readAPIKeyV2Options(
	ctx context.Context, apikeyID string, m interface{},
) (
	jsonAPIKeyV2, error,
) {
	c := m.(*Client)
	var result jsonAPIKeyV2
	body, code, err := c.newRequest(ctx, "/apikeys-v2/"+apikeyID, http.MethodGet, nil)
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

func fillAPIKeyV2(d *schema.ResourceData, jsonData jsonAPIKeyV2) {
	if tfErr := d.Set("apikey_name", jsonData.APIKeyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("profile", jsonData.Profile); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_limitation", jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("apikey", jsonData.APIKey); tfErr != nil {
		panic(tfErr)
	}
}
