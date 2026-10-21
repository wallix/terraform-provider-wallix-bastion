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
	configX509CRLID   = "x509CRLConfig"
	configX509CRLPath = "/config/x509/crl"
)

// crl_auto_fetch is always sent: uploading a crl_file doesn't clear the URL on the Bastion, an
// empty crl_auto_fetch does (also in the same PUT as crl_file). crl_file is write-only, so it is
// only sent on creation and when it changes.
type jsonConfigX509CRL struct {
	CRLAutoFetch string `json:"crl_auto_fetch"`
	CRLCheck     bool   `json:"crl_check"`
	CRLFile      string `json:"crl_file,omitempty"`
}

func resourceConfigX509CRL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigX509CRLCreate,
		ReadContext:   resourceConfigX509CRLRead,
		UpdateContext: resourceConfigX509CRLUpdate,
		DeleteContext: resourceConfigX509CRLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigX509CRLImport,
		},
		Schema: map[string]*schema.Schema{
			skCRLCheck: {
				Type:     schema.TypeBool,
				Required: true,
			},
			skCRLAutoFetch: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{skCRLFile},
			},
			skCRLFile: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{skCRLAutoFetch},
			},
		},
	}
}

func resourceConfigX509CRLVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_config_x509_crl not available with api version %s", version)
}

func resourceConfigX509CRLCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509CRLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigX509CRL(ctx, d, m, true); err != nil {
		return diag.FromErr(err)
	}
	// Since the resource does not have a unique ID, use a static one.
	d.SetId(configX509CRLID)

	return resourceConfigX509CRLRead(ctx, d, m)
}

func resourceConfigX509CRLRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509CRLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigX509CRLOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigX509CRL(d, cfg)

	return nil
}

func resourceConfigX509CRLUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigX509CRLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	// Keep the previous state if the update fails: crl_file is never read back, so a new file
	// saved in state after a failed upload would hide the failure from the next plan.
	d.Partial(true)
	if err := updateConfigX509CRL(ctx, d, m, false); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceConfigX509CRLRead(ctx, d, m)
}

func resourceConfigX509CRLDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// The API has no delete endpoint for this singleton configuration. Disabling the CRL check
	// instead would silently weaken certificate validation, so just drop it from state.
	d.SetId("")

	return nil
}

func resourceConfigX509CRLImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// Since the resource does not have a unique ID, use the static "x509CRLConfig" ID.
	d.SetId(configX509CRLID)

	return []*schema.ResourceData{d}, nil
}

func readConfigX509CRLOptions(
	ctx context.Context, m interface{},
) (jsonConfigX509CRL, error) {
	c := m.(*Client)
	var result jsonConfigX509CRL
	body, code, err := c.newRequest(ctx, configX509CRLPath, http.MethodGet, nil)
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

func updateConfigX509CRL(
	ctx context.Context, d *schema.ResourceData, m interface{}, newResource bool,
) error {
	c := m.(*Client)
	jsonData := prepareConfigX509CRLJSON(d, newResource)
	body, code, err := c.newRequest(ctx, configX509CRLPath, http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// crl_file is only sent on creation and when it changes: the state keeps the last file sent, and
// sending it again on an unrelated update would replace a newer CRL uploaded on the Bastion, such
// as an emergency revocation, without the plan showing it.
func prepareConfigX509CRLJSON(d *schema.ResourceData, newResource bool) jsonConfigX509CRL {
	jsonData := jsonConfigX509CRL{
		CRLAutoFetch: d.Get(skCRLAutoFetch).(string),
		CRLCheck:     d.Get(skCRLCheck).(bool),
	}
	if newResource || d.HasChange(skCRLFile) {
		jsonData.CRLFile = d.Get(skCRLFile).(string)
	}

	return jsonData
}

// crl_file is write-only: the API doesn't return it, so it is kept as configured in state.
func fillConfigX509CRL(d *schema.ResourceData, jsonData jsonConfigX509CRL) {
	if tfErr := d.Set(skCRLAutoFetch, jsonData.CRLAutoFetch); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCRLCheck, jsonData.CRLCheck); tfErr != nil {
		panic(tfErr)
	}
}
