package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCheckoutPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCheckoutPolicyRead,
		Schema: map[string]*schema.Schema{
			"checkout_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skEnableLock: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"change_credentials_at_checkin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"extension": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceCheckoutPolicyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_checkout_policy not available with api version %s", version)
}

func dataSourceCheckoutPolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceCheckoutPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceCheckoutPolicy(ctx, d.Get("checkout_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("checkout_policy_name %s doesn't exists", d.Get("checkout_policy_name").(string)))
	}
	cfg, err := readCheckoutPolicyOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceCheckoutPolicy(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceCheckoutPolicy(d *schema.ResourceData, jsonData jsonCheckoutPolicy) {
	if tfErr := d.Set("checkout_policy_name", jsonData.CheckoutPolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skEnableLock, jsonData.EnableLock); tfErr != nil {
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
