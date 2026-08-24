package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/mod/semver"
)

type jsonCertificateAuthority struct {
	ID                       string `json:"id,omitempty"`
	CertificateAuthorityName string `json:"certificate_authority_name"`
	CAType                   string `json:"ca_type,omitempty"`
	Description              string `json:"description"`
	CACertificate            string `json:"ca_certificate,omitempty"`
}

func resourceCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCertificateAuthorityCreate,
		ReadContext:   resourceCertificateAuthorityRead,
		UpdateContext: resourceCertificateAuthorityUpdate,
		DeleteContext: resourceCertificateAuthorityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCertificateAuthorityImport,
		},
		Schema: map[string]*schema.Schema{
			"certificate_authority_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ca_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{skProtoSSH, "X509"}, false),
			},
			skCACertificate: {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

// certificate_authorities doesn't exist before API v3.12.
func resourceCertificateAuthorityVersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_certificate_authority not available with api version %s", version)
}

func resourceCertificateAuthorityCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceCertificateAuthority(ctx, d.Get("certificate_authority_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf(
			"certificate_authority_name %s already exists", d.Get("certificate_authority_name").(string)))
	}
	id, err := addCertificateAuthority(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceCertificateAuthority(ctx, d.Get("certificate_authority_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"certificate_authority_name %s not found after POST", d.Get("certificate_authority_name").(string)))
		}
	}
	d.SetId(id)

	return resourceCertificateAuthorityRead(ctx, d, m)
}

func resourceCertificateAuthorityRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readCertificateAuthorityOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillCertificateAuthority(d, cfg)
	}

	return nil
}

func resourceCertificateAuthorityUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateCertificateAuthority(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceCertificateAuthorityRead(ctx, d, m)
}

func resourceCertificateAuthorityDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteCertificateAuthority(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCertificateAuthorityImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceCertificateAuthorityVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceCertificateAuthority(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf(
			"don't find certificate_authority_name with id %s (id must be <certificate_authority_name>)", d.Id())
	}
	cfg, err := readCertificateAuthorityOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillCertificateAuthority(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceCertificateAuthority(
	ctx context.Context, certificateAuthorityName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/certificate_authorities/?q=certificate_authority_name="+certificateAuthorityName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonCertificateAuthority
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addCertificateAuthority(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareCertificateAuthorityJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/certificate_authorities/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateCertificateAuthority(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareCertificateAuthorityJSON(d)
	body, code, err := c.newRequest(ctx, "/certificate_authorities/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteCertificateAuthority(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/certificate_authorities/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareCertificateAuthorityJSON(d *schema.ResourceData) jsonCertificateAuthority {
	return jsonCertificateAuthority{
		CertificateAuthorityName: d.Get("certificate_authority_name").(string),
		CAType:                   d.Get("ca_type").(string),
		Description:              d.Get(skDescription).(string),
		CACertificate:            d.Get(skCACertificate).(string),
	}
}

func readCertificateAuthorityOptions(
	ctx context.Context, certificateAuthorityID string, m interface{},
) (
	jsonCertificateAuthority, error,
) {
	c := m.(*Client)
	var result jsonCertificateAuthority
	body, code, err := c.newRequest(ctx, "/certificate_authorities/"+certificateAuthorityID, http.MethodGet, nil)
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

func fillCertificateAuthority(d *schema.ResourceData, jsonData jsonCertificateAuthority) {
	if tfErr := d.Set("certificate_authority_name", jsonData.CertificateAuthorityName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_type", jsonData.CAType); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCACertificate, jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
}
