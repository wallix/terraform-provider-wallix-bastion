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

type jsonEncryption struct {
	Passphrase    string `json:"passphrase,omitempty"`
	NewpassPhrase string `json:"new_passphrase,omitempty"`
}

func resourceEncryption() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEncryptionCreate,
		ReadContext:   resourceEncryptionRead,
		UpdateContext: resourceEncryptionUpdate,
		DeleteContext: resourceEncryptionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceEncryptionImport,
		},
		Schema: map[string]*schema.Schema{
			"current_passphrase": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"new_passphrase": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceEncryptionVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_encryption is not available with API version %s", version)
}

func resourceEncryptionCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceEncryptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}

	// Add encryption
	err := addEncryption(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set a static ID since the API doesn't return one
	d.SetId("encryption")

	// Set persistent attributes
	err = d.Set("new_passphrase", d.Get("new_passphrase").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceEncryptionRead(ctx, d, m)
}

func resourceEncryptionRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceEncryptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	// Verify existence
	exists, err := verifyEncryption(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		// Clear the resource ID if it no longer exists
		d.SetId("")

		return nil
	}
	d.SetId("encryption")
	err = d.Set("new_passphrase", d.Get("new_passphrase").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceEncryptionUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceEncryptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}

	d.Partial(true)

	// Update encryption
	if d.HasChange("current_passphrase") || d.HasChange("new_passphrase") {
		if err := updateEncryption(ctx, d, m); err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("new_passphrase") {
			err := d.Set("new_passphrase", d.Get("new_passphrase"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.Partial(false)

	return resourceEncryptionRead(ctx, d, m)
}

func addEncryption(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareEncryptionJSON(d, false)
	body, code, err := c.newRequest(ctx, "/encryption/", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API didn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateEncryption(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareEncryptionJSON(d, true)
	body, code, err := c.newRequest(ctx, "/encryption", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("API didn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func resourceEncryptionDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// Since the API does not support deletion, we simply remove the resource from the Terraform state
	d.SetId("")

	return nil
}

func verifyEncryption(
	ctx context.Context, m interface{},
) (bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/encryption", http.MethodGet, nil)
	if err != nil {
		return false, err
	}

	if code != http.StatusOK {
		return false, fmt.Errorf("API didn't return OK: %d with body:\n%s", code, body)
	}

	// Check if encryption exists
	var result map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return false, fmt.Errorf("unmarshaling JSON: %w", err)
	}

	// Check if encryption status exists
	var encryptionStatus string
	switch c.bastionAPIVersion {
	case "3.8":
		encryptionStatus, _ = result["encryption"].(string)
		if encryptionStatus != "ready" {
			return false, nil // Return false if encryption is not "ready"
		}
	case "3.12":
		encryptionStatus, _ = result["sealed_state"].(string)
		if encryptionStatus != "unsealed" {
			return false, nil // Return false if encryption is not "unsealed"
		}
	}

	// If encryption is ready (or unsealed for 3.12), return true
	return true, nil
}

func prepareEncryptionJSON(d *schema.ResourceData, update bool) jsonEncryption {
	var jsonData jsonEncryption

	if update {
		jsonData.Passphrase = d.Get("current_passphrase").(string)
	}
	jsonData.NewpassPhrase = d.Get("new_passphrase").(string)

	return jsonData
}

func resourceEncryptionImport(
	d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// For import, assume the static ID
	d.SetId("encryption")

	return []*schema.ResourceData{d}, nil
}
