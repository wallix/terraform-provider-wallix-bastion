package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConnectionPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectionPolicyRead,
		Schema: map[string]*schema.Schema{
			"connection_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_methods": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"options": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectionPolicyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_connection_policy not available with api version %s", version)
}

func dataSourceConnectionPolicyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceConnectionPolicy(ctx, d.Get("connection_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf(
			"connection_policy_name %s doesn't exists", d.Get("connection_policy_name").(string)))
	}
	cfg, err := readConnectionPolicyOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceConnectionPolicy(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceConnectionPolicy(d *schema.ResourceData, jsonData jsonConnectionPolicy) {
	if tfErr := d.Set("connection_policy_name", jsonData.ConnectionPolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("protocol", jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Type != "" {
		if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("type", jsonData.Protocol); tfErr != nil {
			panic(tfErr)
		}
	}
	if tfErr := d.Set("authentication_methods", jsonData.AuthenticationMethods); tfErr != nil {
		panic(tfErr)
	}
	options, _ := json.Marshal(jsonData.Options) //nolint: errchkjson
	if tfErr := d.Set("options", string(options)); tfErr != nil {
		panic(tfErr)
	}
}
