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

type jsonPasswordChangePolicy struct {
	ID                       string  `json:"id,omitempty"`
	PasswordChangePolicyName string  `json:"password_change_policy_name"`
	Description              string  `json:"description"`
	PasswordLength           *int    `json:"password_length,omitempty"`
	SpecialChars             *int    `json:"special_chars,omitempty"`
	LowerChars               *int    `json:"lower_chars,omitempty"`
	UpperChars               *int    `json:"upper_chars,omitempty"`
	DigitChars               *int    `json:"digit_chars,omitempty"`
	ExcludeChars             *string `json:"exclude_chars,omitempty"`
	SSHKeyType               *string `json:"ssh_key_type,omitempty"`
	SSHKeySize               *int    `json:"ssh_key_size,omitempty"`
	ChangePeriod             *string `json:"change_period,omitempty"`
}

func resourcePasswordChangePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordChangePolicyCreate,
		ReadContext:   resourcePasswordChangePolicyRead,
		UpdateContext: resourcePasswordChangePolicyUpdate,
		DeleteContext: resourcePasswordChangePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePasswordChangePolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"password_change_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"special_chars": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"lower_chars": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"upper_chars": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"digit_chars": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"exclude_chars": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ssh_key_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ssh_key_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"change_period": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePasswordChangePolicyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_passwordchangepolicy not available with api version %s", version)
}

func resourcePasswordChangePolicyCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourcePasswordChangePolicy(ctx, d.Get("password_change_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf(
			"password_change_policy_name %s already exists", d.Get("password_change_policy_name").(string)))
	}
	id, err := addPasswordChangePolicy(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourcePasswordChangePolicy(ctx, d.Get("password_change_policy_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"password_change_policy_name %s not found after POST", d.Get("password_change_policy_name").(string)))
		}
	}
	d.SetId(id)

	return resourcePasswordChangePolicyRead(ctx, d, m)
}

func resourcePasswordChangePolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readPasswordChangePolicyOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillPasswordChangePolicy(d, cfg)
	}

	return nil
}

func resourcePasswordChangePolicyUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updatePasswordChangePolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourcePasswordChangePolicyRead(ctx, d, m)
}

func resourcePasswordChangePolicyDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deletePasswordChangePolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePasswordChangePolicyImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourcePasswordChangePolicy(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf(
			"don't find password_change_policy_name with id %s (id must be <password_change_policy_name>)", d.Id())
	}
	cfg, err := readPasswordChangePolicyOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillPasswordChangePolicy(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourcePasswordChangePolicy(
	ctx context.Context, policyName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/passwordchangepolicies/?q=password_change_policy_name="+policyName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonPasswordChangePolicy
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addPasswordChangePolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := preparePasswordChangePolicyJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/passwordchangepolicies/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updatePasswordChangePolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := preparePasswordChangePolicyJSON(d)
	body, code, err := c.newRequest(ctx, "/passwordchangepolicies/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deletePasswordChangePolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/passwordchangepolicies/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// configuredIntPtr returns a pointer to the int value of key if it was explicitly set in the
// user's configuration (including 0), or nil if the attribute was left unconfigured. This
// distinguishes the API's three-state semantics for the *_chars fields (unset => null => char
// class disallowed entirely; 0 => allowed with no minimum; N => minimum of N) from Terraform's
// schema.TypeInt, whose Go zero value (0) would otherwise be indistinguishable from "unset".
func configuredIntPtr(d *schema.ResourceData, key string) *int {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return nil
	}
	attr := raw.GetAttr(key)
	if attr.IsNull() {
		return nil
	}
	v := d.Get(key).(int)

	return &v
}

// configuredStringPtr is the string counterpart of configuredIntPtr, used so an explicitly
// configured empty string (e.g. clearing change_period to deactivate a schedule) is sent to the
// API distinctly from an unconfigured attribute (which leaves the remote value unchanged on
// update, per the API's partial-update semantics).
func configuredStringPtr(d *schema.ResourceData, key string) *string {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return nil
	}
	attr := raw.GetAttr(key)
	if attr.IsNull() {
		return nil
	}
	v := d.Get(key).(string)

	return &v
}

func preparePasswordChangePolicyJSON(d *schema.ResourceData) jsonPasswordChangePolicy {
	return jsonPasswordChangePolicy{
		PasswordChangePolicyName: d.Get("password_change_policy_name").(string),
		Description:              d.Get(skDescription).(string),
		PasswordLength:           configuredIntPtr(d, "password_length"),
		SpecialChars:             configuredIntPtr(d, "special_chars"),
		LowerChars:               configuredIntPtr(d, "lower_chars"),
		UpperChars:               configuredIntPtr(d, "upper_chars"),
		DigitChars:               configuredIntPtr(d, "digit_chars"),
		ExcludeChars:             configuredStringPtr(d, "exclude_chars"),
		SSHKeyType:               configuredStringPtr(d, "ssh_key_type"),
		SSHKeySize:               configuredIntPtr(d, "ssh_key_size"),
		ChangePeriod:             configuredStringPtr(d, "change_period"),
	}
}

func readPasswordChangePolicyOptions(
	ctx context.Context, policyID string, m interface{},
) (
	jsonPasswordChangePolicy, error,
) {
	c := m.(*Client)
	var result jsonPasswordChangePolicy
	body, code, err := c.newRequest(ctx, "/passwordchangepolicies/"+policyID, http.MethodGet, nil)
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

func fillPasswordChangePolicy(d *schema.ResourceData, jsonData jsonPasswordChangePolicy) {
	if tfErr := d.Set("password_change_policy_name", jsonData.PasswordChangePolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_length", intPtrValue(jsonData.PasswordLength)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("special_chars", intPtrValue(jsonData.SpecialChars)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("lower_chars", intPtrValue(jsonData.LowerChars)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("upper_chars", intPtrValue(jsonData.UpperChars)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("digit_chars", intPtrValue(jsonData.DigitChars)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("exclude_chars", stringPtrValue(jsonData.ExcludeChars)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ssh_key_type", stringPtrValue(jsonData.SSHKeyType)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ssh_key_size", intPtrValue(jsonData.SSHKeySize)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("change_period", stringPtrValue(jsonData.ChangePeriod)); tfErr != nil {
		panic(tfErr)
	}
}

func intPtrValue(v *int) int {
	if v == nil {
		return 0
	}

	return *v
}

func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
