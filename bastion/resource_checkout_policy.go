package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonCheckoutPolicy struct {
	ChangeCredentialsAtCheckin bool   `json:"change_credentials_at_checkin"`
	EnableLock                 bool   `json:"enable_lock"`
	Duration                   int    `json:"duration"`
	Extension                  int    `json:"extension"`
	MaxDuration                int    `json:"max_duration"`
	ID                         string `json:"id,omitempty"`
	CheckoutPolicyName         string `json:"checkout_policy_name"`
	Description                string `json:"description"`
}

func resourceCheckoutPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCheckoutPolicyCreate,
		ReadContext:   resourceCheckoutPolicyRead,
		UpdateContext: resourceCheckoutPolicyUpdate,
		DeleteContext: resourceCheckoutPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCheckoutPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"checkout_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_lock": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"duration", "max_duration"},
			},
			"change_credentials_at_checkin": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"enable_lock"},
			},
			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"enable_lock"},
			},
			"extension": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"enable_lock"},
			},
			"max_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"enable_lock"},
			},
		},
	}
}

func resourceCheckoutPolicyVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_checkout_policy not available with api version %s", version)
}

func resourceCheckoutPolicyCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceCheckoutPolicy(ctx, d.Get("checkout_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("checkout_policy_name %s already exists", d.Get("checkout_policy_name").(string)))
	}
	err = addCheckoutPolicy(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceCheckoutPolicy(ctx, d.Get("checkout_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("checkout_policy_name %s not found after POST",
			d.Get("checkout_policy_name").(string)))
	}
	d.SetId(id)

	return resourceCheckoutPolicyRead(ctx, d, m)
}

func resourceCheckoutPolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readCheckoutPolicyOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillCheckoutPolicy(d, cfg)
	}

	return nil
}

func resourceCheckoutPolicyUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateCheckoutPolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceCheckoutPolicyRead(ctx, d, m)
}

func resourceCheckoutPolicyDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteCheckoutPolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCheckoutPolicyImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceCheckoutPolicy(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find checkout_policy_name with id %s (id must be <checkout_policy_name>", d.Id())
	}
	cfg, err := readCheckoutPolicyOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillCheckoutPolicy(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceCheckoutPolicy(
	ctx context.Context, checkoutPolicyName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/checkoutpolicies/?q=checkout_policy_name="+checkoutPolicyName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonCheckoutPolicy
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addCheckoutPolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareCheckoutPolicyJSON(d)
	body, code, err := c.newRequest(ctx, "/checkoutpolicies/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateCheckoutPolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareCheckoutPolicyJSON(d)
	body, code, err := c.newRequest(ctx, "/checkoutpolicies/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteCheckoutPolicy(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/checkoutpolicies/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareCheckoutPolicyJSON(d *schema.ResourceData) jsonCheckoutPolicy {
	var jsonData jsonCheckoutPolicy
	jsonData.CheckoutPolicyName = d.Get("checkout_policy_name").(string)
	jsonData.Description = d.Get("description").(string)
	jsonData.EnableLock = d.Get("enable_lock").(bool)
	jsonData.ChangeCredentialsAtCheckin = d.Get("change_credentials_at_checkin").(bool)
	jsonData.Duration = d.Get("duration").(int)
	jsonData.Extension = d.Get("extension").(int)
	jsonData.MaxDuration = d.Get("max_duration").(int)

	return jsonData
}

func readCheckoutPolicyOptions(
	ctx context.Context, checkoutPolicyID string, m interface{},
) (
	jsonCheckoutPolicy, error,
) {
	c := m.(*Client)
	var result jsonCheckoutPolicy
	body, code, err := c.newRequest(ctx, "/checkoutpolicies/"+checkoutPolicyID, http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillCheckoutPolicy(d *schema.ResourceData, jsonData jsonCheckoutPolicy) {
	if tfErr := d.Set("checkout_policy_name", jsonData.CheckoutPolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable_lock", jsonData.EnableLock); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("change_credentials_at_checkin", jsonData.ChangeCredentialsAtCheckin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("duration", jsonData.Duration); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("extension", jsonData.Extension); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("max_duration", jsonData.MaxDuration); tfErr != nil {
		panic(tfErr)
	}
}
