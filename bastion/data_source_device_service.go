package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeviceService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceServiceRead,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subprotocols": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDeviceServiceVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device_service not available with api version %s", version)
}

func dataSourceDeviceServiceRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceService(ctx, d.Get("device_id").(string), d.Get("service_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("service_name %s on device_id %s doesn't exists",
			d.Get("service_name").(string), d.Get("device_id").(string)))
	}
	cfg, err := readDeviceServiceOptions(ctx, d.Get("device_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceService(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceService(d *schema.ResourceData, jsonData jsonDeviceService) {
	if tfErr := d.Set("service_name", jsonData.ServiceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("connection_policy", jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("protocol", jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("global_domains", jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("subprotocols", jsonData.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
}
