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

type jsonConfigSMTP struct {
	Protocol             string `json:"protocol"`
	AuthenticationMethod string `json:"authentication_method"`
	Server               string `json:"server"`
	Port                 int    `json:"port,omitempty"`
	PostmasterEmail      string `json:"postmaster_email"`
	SenderName           string `json:"sender_name"`
	SenderEmail          string `json:"sender_email"`
	CertificateHash      string `json:"certificate_hash,omitempty"`
	User                 string `json:"user,omitempty"`
	Password             string `json:"password,omitempty"`
}

func resourceConfigSMTP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigSMTPCreate,
		ReadContext:   resourceConfigSMTPRead,
		UpdateContext: resourceConfigSMTPUpdate,
		DeleteContext: resourceConfigSMTPDelete,
		Importer: &schema.ResourceImporter{
			State: resourceConfigSMTPImport,
		},
		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"smtp", "smtps", "starttls"}, false),
			},
			"authentication_method": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"off", "on", "plain", "login", "SCRAM-SHA-1", "CRAM-MD5", "DIGEST-MD5", "ntlm",
				}, false),
			},
			"server": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      25,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"postmaster_email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sender_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sender_email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceConfigSMTPVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_config_smtp not available with api version %s", version)
}

func resourceConfigSMTPCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigSMTPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigSMTP(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	// Since the resource does not have a unique ID, use a static one.
	d.SetId("smtpConfig")

	return resourceConfigSMTPRead(ctx, d, m)
}

func resourceConfigSMTPRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigSMTPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigSMTPOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigSMTP(d, cfg)

	return nil
}

func resourceConfigSMTPUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigSMTPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigSMTP(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceConfigSMTPRead(ctx, d, m)
}

func resourceConfigSMTPDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// The API has no delete endpoint for this singleton configuration; just drop it from state.
	d.SetId("")

	return nil
}

func resourceConfigSMTPImport(
	d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// Since the resource does not have a unique ID, use the static "smtpConfig" ID.
	d.SetId("smtpConfig")

	return []*schema.ResourceData{d}, nil
}

func readConfigSMTPOptions(
	ctx context.Context, m interface{},
) (jsonConfigSMTP, error) {
	c := m.(*Client)
	var result jsonConfigSMTP
	body, code, err := c.newRequest(ctx, "/config/smtp", http.MethodGet, nil)
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

func updateConfigSMTP(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareConfigSMTPJSON(d)
	body, code, err := c.newRequest(ctx, "/config/smtp", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareConfigSMTPJSON(d *schema.ResourceData) jsonConfigSMTP {
	return jsonConfigSMTP{
		Protocol:             d.Get("protocol").(string),
		AuthenticationMethod: d.Get("authentication_method").(string),
		Server:               d.Get("server").(string),
		Port:                 d.Get("port").(int),
		PostmasterEmail:      d.Get("postmaster_email").(string),
		SenderName:           d.Get("sender_name").(string),
		SenderEmail:          d.Get("sender_email").(string),
		CertificateHash:      d.Get("certificate_hash").(string),
		User:                 d.Get("user").(string),
		Password:             d.Get("password").(string),
	}
}

func fillConfigSMTP(d *schema.ResourceData, jsonData jsonConfigSMTP) {
	if tfErr := d.Set("protocol", jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authentication_method", jsonData.AuthenticationMethod); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("server", jsonData.Server); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("postmaster_email", jsonData.PostmasterEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sender_name", jsonData.SenderName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("sender_email", jsonData.SenderEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_hash", jsonData.CertificateHash); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user", jsonData.User); tfErr != nil {
		panic(tfErr)
	}
}
