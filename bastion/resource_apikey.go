package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type jsonAPIKey struct {
	ID           string `json:"id,omitempty"`
	APIKeyName   string `json:"apikey_name"`
	APIKey       string `json:"apikey,omitempty"`
	IPLimitation string `json:"ip_limitation"`
}

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIKeyImport,
		},
		Schema: map[string]*schema.Schema{
			"apikey_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skIPLimitation: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"apikey": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAPIKeyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_apikey not available with api version %s", version)
}

func resourceAPIKeyCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAPIKey(ctx, d.Get("apikey_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("apikey_name %s already exists", d.Get("apikey_name").(string)))
	}
	id, err := addAPIKey(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceAPIKey(ctx, d.Get("apikey_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("apikey_name %s not found after POST", d.Get("apikey_name").(string)))
		}
	}
	d.SetId(id)

	return resourceAPIKeyRead(ctx, d, m)
}

func resourceAPIKeyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAPIKeyOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAPIKey(d, cfg)
	}

	return nil
}

func resourceAPIKeyUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAPIKey(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAPIKeyRead(ctx, d, m)
}

func resourceAPIKeyDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAPIKey(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAPIKeyImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAPIKey(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find apikey_name with id %s (id must be <apikey_name>)", d.Id())
	}
	cfg, err := readAPIKeyOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillAPIKey(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAPIKey(
	ctx context.Context, apikeyName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/apikeys/?q=apikey_name="+apikeyName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAPIKey
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAPIKey(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareAPIKeyJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/apikeys/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateAPIKey(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAPIKeyJSON(d)
	body, code, err := c.newRequest(ctx, "/apikeys/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAPIKey(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/apikeys/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareAPIKeyJSON(d *schema.ResourceData) jsonAPIKey {
	return jsonAPIKey{
		APIKeyName:   d.Get("apikey_name").(string),
		IPLimitation: d.Get(skIPLimitation).(string),
	}
}

func readAPIKeyOptions(
	ctx context.Context, apikeyID string, m interface{},
) (
	jsonAPIKey, error,
) {
	c := m.(*Client)
	var result jsonAPIKey
	body, code, err := c.newRequest(ctx, "/apikeys/"+apikeyID, http.MethodGet, nil)
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

func fillAPIKey(d *schema.ResourceData, jsonData jsonAPIKey) {
	if tfErr := d.Set("apikey_name", jsonData.APIKeyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skIPLimitation, jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("apikey", jsonData.APIKey); tfErr != nil {
		panic(tfErr)
	}
}
