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

// NewPassphrase is always sent: an empty one sets up the Bastion without a passphrase. Passphrase,
// the current one, is only sent to change it.
type jsonEncryption struct {
	Passphrase    *string `json:"passphrase,omitempty"`
	NewPassphrase string  `json:"new_passphrase"`
}

// jsonEncryptionState is the GET response, with the field names of API 3.8 and later.
type jsonEncryptionState struct {
	SealState      string `json:"seal_state"`
	EncryptionMode string `json:"encryption_mode"`
}

const (
	encryptionNeedSetup = "need_setup"
	encryptionSealed    = "sealed"
	encryptionUnsealed  = "unsealed"
)

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

	// A setup request is only for a Bastion that isn't set up: on one that is, changing the
	// passphrase requires the current one, through an import and an update.
	state, err := readEncryptionState(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if state.SealState != encryptionNeedSetup {
		return diag.Errorf("the Bastion encryption is already set up (seal_state %q): import it with "+
			"`terraform import wallix-bastion_encryption.<name> encryption`", state.SealState)
	}
	// Add encryption
	err = addEncryption(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set a static ID since the API doesn't return one
	d.SetId("encryption")

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
	switch state.SealState {
	case encryptionNeedSetup:
		// Not set up (fresh install): the next apply sets it up.
		d.SetId("")

		return nil
	case encryptionSealed:
		// Dropping it from state would make the next apply send a setup request to a sealed
		// Bastion: it must be unsealed with its passphrase first.
		return diag.Errorf("the Bastion encryption is sealed: unseal it with its passphrase before " +
			"Terraform can manage wallix-bastion_encryption")
	case encryptionUnsealed:
	default:
		return diag.Errorf("unknown seal_state %q of the Bastion encryption", state.SealState)
	}
	d.SetId("encryption")

	return nil
}

func resourceEncryptionUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceEncryptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	// Only a new passphrase changes something on the Bastion: updating current_passphrase alone,
	// for example after a change, must not send another change request.
	if d.HasChange("new_passphrase") {
		// Keep the previous state if the change fails: recorded, the new passphrase would look
		// applied while the Bastion keeps the old one, and the next plan would be empty.
		d.Partial(true)
		if err := updateEncryption(ctx, d, m); err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
	}

	return resourceEncryptionRead(ctx, d, m)
}

func addEncryption(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareEncryptionJSON(d, false)
	body, code, err := c.newRequest(ctx, "/encryption", http.MethodPut, jsonData)
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
		return state, fmt.Errorf("API didn't return OK: %d with body:\n%s", code, body)
	}
	if err := json.Unmarshal([]byte(body), &state); err != nil {
		return state, fmt.Errorf("unmarshaling JSON: %w", err)
	}

	return state, nil
}

func prepareEncryptionJSON(d *schema.ResourceData, update bool) jsonEncryption {
	jsonData := jsonEncryption{NewPassphrase: d.Get("new_passphrase").(string)}
	if update {
		current := d.Get("current_passphrase").(string)
		jsonData.Passphrase = &current
	}

	return jsonData
}

func resourceEncryptionImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	// For import, assume the static ID
	d.SetId("encryption")

	return []*schema.ResourceData{d}, nil
}
