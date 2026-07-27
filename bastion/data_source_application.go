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
			"connection_policy": {
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_domains": {
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
						"target": {
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
			"target": {
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
						"admin_account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_password_change": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"password_change_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_change_plugin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_change_plugin_parameters": {
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
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
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
	if tfErr := d.Set("connection_policy", jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	category := jsonData.Category
	if category == "" {
		category = "standard"
	}
	if tfErr := d.Set("category", category); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, "application_url", jsonData.ApplicationURL)
	setApplicationOptionalString(d, "browser", jsonData.Browser)
	setApplicationOptionalString(d, "browser_version", jsonData.BrowserVersion)
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("global_domains", jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("parameters", jsonData.Parameters); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("paths", fillApplicationPaths(jsonData.Paths)); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, "target", jsonData.Target)
	if tfErr := d.Set("local_domains", fillApplicationLocalDomains(jsonData.LocalDomains)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("tags", fillApplicationTags(jsonData.Tags)); tfErr != nil {
		panic(tfErr)
	}
}
