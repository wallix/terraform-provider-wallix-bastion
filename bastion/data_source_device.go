package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceRead,
		Schema: map[string]*schema.Schema{
			"device_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
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
						skDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skAdminAccount: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skCAPublicKey: {
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
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						skServiceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skConnectionPolicy: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						skProtocol: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skGlobalDomains: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						skSubprotocols: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

func dataSourceDeviceVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device not available with api version %s", version)
}

func dataSourceDeviceRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDevice(ctx, d.Get("device_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("device_name %s doesn't exists", d.Get("device_name").(string)))
	}
	cfg, err := readDeviceOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDevice(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDevice(d *schema.ResourceData, jsonData jsonDevice) {
	if tfErr := d.Set("device_name", jsonData.DeviceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("alias", jsonData.Alias); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	localDomains := make([]map[string]interface{}, 0)
	if jsonData.LocalDomains != nil {
		localDomains = make([]map[string]interface{}, len(*jsonData.LocalDomains))
		for i, v := range *jsonData.LocalDomains {
			localDomains[i] = map[string]interface{}{
				"id":                   v.ID,
				skAdminAccount:         v.AdminAccount,
				skDomainName:           v.DomainName,
				skCAPublicKey:          v.CAPublicKey,
				skDescription:          v.Description,
				skEnablePasswordChange: v.EnablePasswordChange,
				skPasswordChangePolicy: v.PasswordChangePolicy,
				skPasswordChangePlugin: v.PasswordChangePlugin,
			}
			pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters) //nolint: errchkjson
			localDomains[i][skPasswordChangePluginParameters] = string(pluginParameters)
		}
	}
	if tfErr := d.Set("local_domains", localDomains); tfErr != nil {
		panic(tfErr)
	}
	services := make([]map[string]interface{}, 0)
	if jsonData.Services != nil {
		services = make([]map[string]interface{}, len(*jsonData.Services))
		for i, v := range *jsonData.Services {
			service := map[string]interface{}{
				"id":               v.ID,
				skServiceName:      v.ServiceName,
				skConnectionPolicy: v.ConnectionPolicy,
				skPort:             v.Port,
				skProtocol:         v.Protocol,
				skGlobalDomains:    make([]string, 0),
				skSubprotocols:     make([]string, 0),
			}
			if v.GlobalDomains != nil {
				service[skGlobalDomains] = make(([]string), len(*v.GlobalDomains))
				copy(service[skGlobalDomains].([]string), *v.GlobalDomains)
			}
			if v.SubProtocols != nil {
				service[skSubprotocols] = make(([]string), len(*v.SubProtocols))
				copy(service[skSubprotocols].([]string), *v.SubProtocols)
			}
			services[i] = service
		}
	}
	if tfErr := d.Set("services", services); tfErr != nil {
		panic(tfErr)
	}

	if jsonData.Tags != nil && len(*jsonData.Tags) > 0 {
		apiTags := *jsonData.Tags

		stateTags := make([]interface{}, len(apiTags))

		for i, tagMap := range apiTags {
			stateMap := map[string]interface{}{
				skKey:   tagMap[skKey],
				skValue: tagMap[skValue],
			}
			stateTags[i] = stateMap
		}

		if tfErr := d.Set("tags", stateTags); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("tags", make([]interface{}, 0)); tfErr != nil {
			panic(tfErr)
		}
	}
}
