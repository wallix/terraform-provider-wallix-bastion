package bastion

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	sleepTimeAfterX509ConfigChange = 3 * time.Second
)

type jsonConfigX509 struct {
	CaCertificate    string `json:"ca_certificate,omitempty"`
	ServerPublicKey  string `json:"server_public_key"`
	ServerPrivateKey string `json:"server_private_key"`
	Enable           bool   `json:"enable"`
	Default          bool   `json:"default,omitempty"`
}

func resourceConfigX509() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigX509Create,
		ReadContext:   resourceConfigX509Read,
		UpdateContext: resourceConfigX509Update,
		DeleteContext: resourceConfigX509Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigX509Import,
		},
		Schema: map[string]*schema.Schema{
			"ca_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_public_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_private_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceConfigX509Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Add the configuration
	if err := addConfigX509(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	// Use a static ID since the API does not provide one
	d.SetId("x509Config")

	return resourceConfigX509Read(ctx, d, m)
}

func resourceConfigX509Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cfg, err := readConfigX509Options(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// If default config, mark the resource as deleted
	if cfg.Default {
		d.SetId("")

		return nil
	}
	if d.Get("ca_certificate").(string) != "" {
		// check diff between api response and common name of ca_certificate
		caCertificatePEM, _ := pem.Decode([]byte(d.Get("ca_certificate").(string)))
		if caCertificatePEM == nil {
			return diag.FromErr(errors.New("failed to decode PEM block from ca_certificate"))
		}
		caCertificate, err := x509.ParseCertificate(caCertificatePEM.Bytes)
		if err != nil {
			return diag.FromErr(err)
		}
		// If ca_certificate common name not match, mark the resource as deleted
		if !strings.Contains(cfg.CaCertificate, "/CN="+caCertificate.Subject.CommonName) {
			d.SetId("")

			return nil
		}
	}
	if d.Get("server_public_key").(string) != "" {
		// check diff between api response and common name of server_public_key
		serverPublicKeyPEM, _ := pem.Decode([]byte(d.Get("server_public_key").(string)))
		if serverPublicKeyPEM == nil {
			return diag.FromErr(errors.New("failed to decode PEM block from server_public_key"))
		}
		serverPublicKey, err := x509.ParseCertificate(serverPublicKeyPEM.Bytes)
		if err != nil {
			return diag.FromErr(err)
		}
		// If server_public_key common name not match, mark the resource as deleted
		if !strings.Contains(cfg.ServerPublicKey, "/CN="+serverPublicKey.Subject.CommonName) {
			d.SetId("")

			return nil
		}
	}

	if err := fillConfigX509(d, cfg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConfigX509Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := updateConfigX509(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceConfigX509Read(ctx, d, m)
}

func resourceConfigX509Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := deleteConfigX509(ctx, m); err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource from state
	d.SetId("")

	return nil
}

func resourceConfigX509Import(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// Since the resource does not have a unique ID, use the static "x509Config" ID
	d.SetId("x509Config")

	return []*schema.ResourceData{d}, nil
}

func addConfigX509(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareConfigX509JSON(d)
	body, code, err := c.newRequest(ctx, "/config/x509", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API returned error: %d with body:\n%s", code, body)
	}

	// sleep after modifying the x509 configuration
	// to wait for the API listener to restart with the new certificate
	time.Sleep(sleepTimeAfterX509ConfigChange)

	return nil
}

func readConfigX509Options(ctx context.Context, m interface{}) (jsonConfigX509, error) {
	c := m.(*Client)
	var result jsonConfigX509
	body, code, err := c.newRequest(ctx, "/config/x509", http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("API returned error: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return result, nil
}

func updateConfigX509(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareConfigX509JSON(d)
	body, code, err := c.newRequest(ctx, "/config/x509", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API returned error: %d with body:\n%s", code, body)
	}

	// sleep after modifying the x509 configuration
	// to wait for the API listener to restart with the new certificate
	time.Sleep(sleepTimeAfterX509ConfigChange)

	return nil
}

func deleteConfigX509(ctx context.Context, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/config/x509", http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API returned error: %d with body:\n%s", code, body)
	}

	// sleep after modifying the x509 configuration
	// to wait for the API listener to restart with the new certificate
	time.Sleep(sleepTimeAfterX509ConfigChange)

	return nil
}

func prepareConfigX509JSON(d *schema.ResourceData) jsonConfigX509 {
	return jsonConfigX509{
		CaCertificate:    d.Get("ca_certificate").(string),
		ServerPublicKey:  d.Get("server_public_key").(string),
		ServerPrivateKey: d.Get("server_private_key").(string),
		Enable:           d.Get("enable").(bool),
	}
}

//nolint:wrapcheck
func fillConfigX509(d *schema.ResourceData, jsonData jsonConfigX509) error {
	if _, enableExplicitlySet := d.GetOk("enable"); enableExplicitlySet || jsonData.Enable {
		if err := d.Set("enable", jsonData.Enable); err != nil {
			return err
		}
	}

	return nil
}
