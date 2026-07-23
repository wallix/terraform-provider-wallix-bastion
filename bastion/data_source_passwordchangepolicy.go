package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePasswordChangePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePasswordChangePolicyRead,
		Schema: map[string]*schema.Schema{
			"password_change_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"special_chars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"lower_chars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"upper_chars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"digit_chars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"exclude_chars": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_key_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"change_period": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePasswordChangePolicyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_passwordchangepolicy not available with api version %s", version)
}

func dataSourcePasswordChangePolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourcePasswordChangePolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourcePasswordChangePolicy(ctx, d.Get("password_change_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf(
			"password_change_policy_name %s doesn't exists", d.Get("password_change_policy_name").(string)))
	}
	cfg, err := readPasswordChangePolicyOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourcePasswordChangePolicy(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourcePasswordChangePolicy(d *schema.ResourceData, jsonData jsonPasswordChangePolicy) {
	if tfErr := d.Set("password_change_policy_name", jsonData.PasswordChangePolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
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
