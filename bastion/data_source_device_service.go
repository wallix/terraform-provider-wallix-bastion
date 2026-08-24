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
			skDeviceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skServiceName: {
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skSubprotocols: {
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
	id, ex, err := searchResourceDeviceService(ctx, d.Get(skDeviceID).(string), d.Get(skServiceName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("service_name %s on device_id %s doesn't exists",
			d.Get(skServiceName).(string), d.Get(skDeviceID).(string)))
	}
	cfg, err := readDeviceServiceOptions(ctx, d.Get(skDeviceID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceService(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceService(d *schema.ResourceData, jsonData jsonDeviceService) {
	if tfErr := d.Set(skServiceName, jsonData.ServiceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skConnectionPolicy, jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skProtocol, jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skGlobalDomains, jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSubprotocols, jsonData.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
}
