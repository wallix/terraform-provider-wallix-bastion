package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccounts: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skAccountMappings: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skInteractiveLogins: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClusterVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_cluster not available with api version %s", version)
}

func dataSourceClusterRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceCluster(ctx, d.Get("cluster_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("cluster_name %s doesn't exists", d.Get("cluster_name").(string)))
	}
	cfg, err := readClusterOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceCluster(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceCluster(d *schema.ResourceData, jsonData jsonCluster) {
	if tfErr := d.Set("cluster_name", jsonData.ClusterName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccounts, jsonData.Accounts); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountMappings, jsonData.AccountMappings); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skInteractiveLogins, jsonData.InteractiveLogins); tfErr != nil {
		panic(tfErr)
	}
}
