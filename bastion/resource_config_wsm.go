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

type jsonConfigWSM struct {
	Hostname   *string `json:"hostname"`
	JwsPrivate string  `json:"jws_private,omitempty"`
	JwsPublic  string  `json:"jws_public,omitempty"`
	JwePublic  *string `json:"jwe_public"`
}

func resourceConfigWSM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigWSMCreate,
		ReadContext:   resourceConfigWSMRead,
		UpdateContext: resourceConfigWSMUpdate,
		DeleteContext: resourceConfigWSMDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigWSMImport,
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"jwe_public": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"jws_private": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"jws_public": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// config/wsm doesn't exist before API v3.12.
func resourceConfigWSMVersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_config_wsm not available with api version %s", version)
}

func resourceConfigWSMCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigWSMVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigWSM(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	// Since the resource does not have a unique ID, use a static one.
	d.SetId("wsmConfig")

	return resourceConfigWSMRead(ctx, d, m)
}

func resourceConfigWSMRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigWSMVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigWSMOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigWSM(d, cfg)

	return nil
}

func resourceConfigWSMUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigWSMVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigWSM(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceConfigWSMRead(ctx, d, m)
}

func resourceConfigWSMDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// The API has no delete endpoint for this singleton configuration; just drop it from state.
	d.SetId("")

	return nil
}

func resourceConfigWSMImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// Since the resource does not have a unique ID, use the static "wsmConfig" ID.
	d.SetId("wsmConfig")

	return []*schema.ResourceData{d}, nil
}

func readConfigWSMOptions(
	ctx context.Context, m interface{},
) (jsonConfigWSM, error) {
	c := m.(*Client)
	var result jsonConfigWSM
	body, code, err := c.newRequest(ctx, "/config/wsm", http.MethodGet, nil)
	if err != nil {
		return result, err
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

func updateConfigWSM(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareConfigWSMJSON(d)
	body, code, err := c.newRequest(ctx, "/config/wsm", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// prepareConfigWSMJSON sends an explicit JSON null for hostname/jwe_public when unset - the API
// rejects an empty string ("hostname: A valid format is required.", confirmed live) but accepts
// null to clear the field.
func prepareConfigWSMJSON(d *schema.ResourceData) jsonConfigWSM {
	jsonData := jsonConfigWSM{}
	if hostname := d.Get("hostname").(string); hostname != "" {
		jsonData.Hostname = &hostname
	}
	if jwePublic := d.Get("jwe_public").(string); jwePublic != "" {
		jsonData.JwePublic = &jwePublic
	}

	return jsonData
}

func fillConfigWSM(d *schema.ResourceData, jsonData jsonConfigWSM) {
	hostname := ""
	if jsonData.Hostname != nil {
		hostname = *jsonData.Hostname
	}
	if tfErr := d.Set("hostname", hostname); tfErr != nil {
		panic(tfErr)
	}
	jwePublic := ""
	if jsonData.JwePublic != nil {
		jwePublic = *jsonData.JwePublic
	}
	if tfErr := d.Set("jwe_public", jwePublic); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("jws_private", jsonData.JwsPrivate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("jws_public", jsonData.JwsPublic); tfErr != nil {
		panic(tfErr)
	}
}
