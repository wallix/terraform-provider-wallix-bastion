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
			"host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"admin_account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ca_public_key": {
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
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"connection_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"global_domains": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subprotocols": {
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
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("alias", jsonData.Alias); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	localDomains := make([]map[string]interface{}, 0)
	if jsonData.LocalDomains != nil {
		localDomains = make([]map[string]interface{}, len(*jsonData.LocalDomains))
		for i, v := range *jsonData.LocalDomains {
			localDomains[i] = map[string]interface{}{
				"id":                     v.ID,
				"admin_account":          v.AdminAccount,
				"domain_name":            v.DomainName,
				"ca_public_key":          v.CAPublicKey,
				"description":            v.Description,
				"enable_password_change": v.EnablePasswordChange,
				"password_change_policy": v.PasswordChangePolicy,
				"password_change_plugin": v.PasswordChangePlugin,
			}
			pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters) //nolint: errchkjson
			localDomains[i]["password_change_plugin_parameters"] = string(pluginParameters)
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
				"id":                v.ID,
				"service_name":      v.ServiceName,
				"connection_policy": v.ConnectionPolicy,
				"port":              v.Port,
				"protocol":          v.Protocol,
				"global_domains":    make([]string, 0),
				"subprotocols":      make([]string, 0),
			}
			if v.GlobalDomains != nil {
				service["global_domains"] = make(([]string), len(*v.GlobalDomains))
				copy(service["global_domains"].([]string), *v.GlobalDomains)
			}
			if v.SubProtocols != nil {
				service["subprotocols"] = make(([]string), len(*v.SubProtocols))
				copy(service["subprotocols"].([]string), *v.SubProtocols)
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
				"key":   tagMap["key"],
				"value": tagMap["value"],
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
