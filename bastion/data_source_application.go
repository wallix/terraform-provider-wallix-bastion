package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationRead,
		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skConnectionPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"browser": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"browser_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skGlobalDomains: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"paths": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						skTarget: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"program": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"working_dir": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			skTarget: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						skAdminAccount: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skEnablePasswordChange: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						skPasswordChangePolicy: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPasswordChangePlugin: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPasswordChangePluginParameters: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						skKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceApplicationVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_application not available with api version %s", version)
}

func dataSourceApplicationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplication(ctx, d.Get("application_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("application_name %s doesn't exists", d.Get("application_name").(string)))
	}
	cfg, err := readApplicationOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplication(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplication(d *schema.ResourceData, jsonData jsonApplication) {
	if tfErr := d.Set("application_name", jsonData.ApplicationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skConnectionPolicy, jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	category := jsonData.Category
	if category == "" {
		category = skStandard
	}
	if tfErr := d.Set("category", category); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, "application_url", jsonData.ApplicationURL)
	setApplicationOptionalString(d, "browser", jsonData.Browser)
	setApplicationOptionalString(d, "browser_version", jsonData.BrowserVersion)
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skGlobalDomains, jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("parameters", jsonData.Parameters); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("paths", fillApplicationPaths(jsonData.Paths)); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, skTarget, jsonData.Target)
	if tfErr := d.Set("local_domains", fillApplicationLocalDomains(jsonData.LocalDomains)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("tags", fillApplicationTags(jsonData.Tags)); tfErr != nil {
		panic(tfErr)
	}
}
