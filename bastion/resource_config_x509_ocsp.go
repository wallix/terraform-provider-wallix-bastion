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

const (
	configX509OCSPID   = "x509OCSPConfig"
	configX509OCSPPath = "/config/x509/ocsp"
)

type jsonConfigX509OCSP struct {
	OCSPCheck             bool   `json:"ocsp_check"`
	DefaultOCSPResponder  string `json:"default_ocsp_responder"`
	OverrideOCSPResponder bool   `json:"override_ocsp_responder"`
	OCSPProxy             string `json:"ocsp_proxy"`
}

func resourceConfigX509OCSP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigX509OCSPCreate,
		ReadContext:   resourceConfigX509OCSPRead,
		UpdateContext: resourceConfigX509OCSPUpdate,
		DeleteContext: resourceConfigX509OCSPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigX509OCSPImport,
		},
		Schema: map[string]*schema.Schema{
			skOCSPCheck: {
				Type:     schema.TypeBool,
				Required: true,
			},
			skDefaultOCSPResponder: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skOverrideOCSPResponder: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// The proxy URL can hold credentials (http://user:pass@proxy:3128), which the API returns.
			skOCSPProxy: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceConfigX509OCSPVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_config_x509_ocsp not available with api version %s", version)
}

func resourceConfigX509OCSPCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509OCSPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigX509OCSP(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	// Since the resource does not have a unique ID, use a static one.
	d.SetId(configX509OCSPID)

	return resourceConfigX509OCSPRead(ctx, d, m)
}

func resourceConfigX509OCSPRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509OCSPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigX509OCSPOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigX509OCSP(d, cfg)

	return nil
}

func resourceConfigX509OCSPUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509OCSPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigX509OCSP(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceConfigX509OCSPRead(ctx, d, m)
}

func resourceConfigX509OCSPDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// The API has no delete endpoint for this singleton configuration. Disabling the OCSP check
	// instead would silently weaken certificate validation, so just drop it from state.
	d.SetId("")

	return nil
}

func resourceConfigX509OCSPImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// Since the resource does not have a unique ID, use the static "x509OCSPConfig" ID.
	d.SetId(configX509OCSPID)

	return []*schema.ResourceData{d}, nil
}

func readConfigX509OCSPOptions(
	ctx context.Context, m interface{},
) (jsonConfigX509OCSP, error) {
	c := m.(*Client)
	var result jsonConfigX509OCSP
	body, code, err := c.newRequest(ctx, configX509OCSPPath, http.MethodGet, nil)
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

func updateConfigX509OCSP(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareConfigX509OCSPJSON(d)
	body, code, err := c.newRequest(ctx, configX509OCSPPath, http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareConfigX509OCSPJSON(d *schema.ResourceData) jsonConfigX509OCSP {
	return jsonConfigX509OCSP{
		OCSPCheck:             d.Get(skOCSPCheck).(bool),
		DefaultOCSPResponder:  d.Get(skDefaultOCSPResponder).(string),
		OverrideOCSPResponder: d.Get(skOverrideOCSPResponder).(bool),
		OCSPProxy:             d.Get(skOCSPProxy).(string),
	}
}

func fillConfigX509OCSP(d *schema.ResourceData, jsonData jsonConfigX509OCSP) {
	if tfErr := d.Set(skOCSPCheck, jsonData.OCSPCheck); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultOCSPResponder, jsonData.DefaultOCSPResponder); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skOverrideOCSPResponder, jsonData.OverrideOCSPResponder); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skOCSPProxy, jsonData.OCSPProxy); tfErr != nil {
		panic(tfErr)
	}
}
