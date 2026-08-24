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

type jsonExternalAuthRadius struct {
	Port                 int     `json:"port"`
	Timeout              float64 `json:"timeout"`
	ID                   string  `json:"id,omitempty"`
	AuthenticationName   string  `json:"authentication_name"`
	Description          string  `json:"description"`
	Host                 string  `json:"host"`
	Secret               string  `json:"secret"`
	Type                 string  `json:"type"`
	UsePrimaryAuthDomain bool    `json:"use_primary_auth_domain"`
}

func resourceExternalAuthRadius() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalAuthRadiusCreate,
		ReadContext:   resourceExternalAuthRadiusRead,
		UpdateContext: resourceExternalAuthRadiusUpdate,
		DeleteContext: resourceExternalAuthRadiusDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceExternalAuthRadiusImport,
		},
		Schema: map[string]*schema.Schema{
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Required: true,
			},
			skPort: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			skTimeout: {
				Type:     schema.TypeFloat,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skUsePrimaryAuthDomain: {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceExternalAuthRadiusVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_externalauth_radius not available with api version %s", version)
}

func resourceExternalAuthRadiusCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceExternalAuthRadius(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s already exists", d.Get(skAuthenticationName).(string)))
	}
	id, err := addExternalAuthRadius(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceExternalAuthRadius(ctx, d.Get(skAuthenticationName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("authentication_name %s not found after POST", d.Get(skAuthenticationName).(string)))
		}
	}
	d.SetId(id)

	return resourceExternalAuthRadiusRead(ctx, d, m)
}

func resourceExternalAuthRadiusRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readExternalAuthRadiusOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillExternalAuthRadius(d, cfg)
	}

	return nil
}

func resourceExternalAuthRadiusUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateExternalAuthRadius(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceExternalAuthRadiusRead(ctx, d, m)
}

func resourceExternalAuthRadiusDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteExternalAuthRadius(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceExternalAuthRadiusImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceExternalAuthRadius(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authentication_name with id %s (id must be <authentication_name>)", d.Id())
	}
	cfg, err := readExternalAuthRadiusOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillExternalAuthRadius(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceExternalAuthRadius(
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
	var results []jsonExternalAuthRadius
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addExternalAuthRadius(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareExternalAuthRadiusJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateExternalAuthRadius(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareExternalAuthRadiusJSON(d)
	body, code, err := c.newRequest(ctx, "/externalauths/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteExternalAuthRadius(
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

func prepareExternalAuthRadiusJSON(d *schema.ResourceData) jsonExternalAuthRadius {
	return jsonExternalAuthRadius{
		AuthenticationName:   d.Get(skAuthenticationName).(string),
		Host:                 d.Get(skHost).(string),
		Port:                 d.Get(skPort).(int),
		Secret:               d.Get("secret").(string),
		Timeout:              d.Get(skTimeout).(float64),
		Description:          d.Get(skDescription).(string),
		UsePrimaryAuthDomain: d.Get(skUsePrimaryAuthDomain).(bool),
		Type:                 "RADIUS",
	}
}

func readExternalAuthRadiusOptions(
	ctx context.Context, authenticationID string, m interface{},
) (
	jsonExternalAuthRadius, error,
) {
	c := m.(*Client)
	var result jsonExternalAuthRadius
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

func fillExternalAuthRadius(d *schema.ResourceData, jsonData jsonExternalAuthRadius) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skTimeout, jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
