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
			StateContext: resourceEncryptionImport,
		},
		Schema: map[string]*schema.Schema{
			"current_passphrase": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"new_passphrase": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
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
	state, err := readEncryptionState(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	// Encryption that was never set up doesn't exist yet. A sealed Bastion (after a reboot, until
	// the passphrase is entered) still has its encryption set up: keep it in state, or the next
	// plan would propose to set the encryption up again.
	if state.SealState == encryptionNeedSetup {
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

func prepareEncryptionJSON(d *schema.ResourceData, update bool) jsonEncryption {
	var jsonData jsonEncryption

	if update {
		jsonData.Passphrase = d.Get("current_passphrase").(string)
	}
	jsonData.NewpassPhrase = d.Get("new_passphrase").(string)

	return jsonData
}

func resourceEncryptionImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// For import, assume the static ID
	d.SetId("encryption")

	return []*schema.ResourceData{d}, nil
}

const (
	encryptionNeedSetup = "need_setup"
	encryptionSealed    = "sealed"
	encryptionUnsealed  = "unsealed"
)

// jsonEncryptionState is the response of GET /encryption. API 3.8 renamed the fields "encryption"
// to "seal_state" and "security_level" to "encryption_mode", so both supported API versions (3.8
// and 3.12) return these ones (checked on a Bastion 12.4).
type jsonEncryptionState struct {
	SealState      string `json:"seal_state"`
	EncryptionMode string `json:"encryption_mode"`
}

// readEncryptionState returns the encryption state of the Bastion. seal_state is need_setup
// (encryption never set up), sealed (set up, waiting for the passphrase after a reboot) or
// unsealed (ready for use).
func readEncryptionState(
	ctx context.Context, m interface{},
) (jsonEncryptionState, error) {
	c := m.(*Client)
	var state jsonEncryptionState
	body, code, err := c.newRequest(ctx, "/encryption", http.MethodGet, nil)
	if err != nil {
		return state, err
	}
	if code != http.StatusOK {
		return state, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	if err := json.Unmarshal([]byte(body), &state); err != nil {
		return state, fmt.Errorf("unmarshaling json: %w", err)
	}
	switch state.SealState {
	case encryptionNeedSetup, encryptionSealed, encryptionUnsealed:
		return state, nil
	default:
		return state, fmt.Errorf("unexpected seal_state %q in the encryption state:\n%s", state.SealState, body)
	}
}
