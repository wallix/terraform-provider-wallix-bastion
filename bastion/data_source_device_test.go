package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDevice_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDeviceConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDeviceConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_device.testacc_dataDevice",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device.testacc_dataDevice",
						"device_name", "testacc_Device_ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device.testacc_dataDevice",
						"host", "testacc-ds.device"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device.testacc_dataDevice",
						"local_domains.#", "0"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device.testacc_dataDevice",
						"services.#", "0"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device.testacc_dataDevice",
						"tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.wallix-bastion_device.testacc_dataDevice",
						"tags.*", map[string]string{
							"key":   "testkey",
							"value": "testvalue",
						}),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDeviceConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_Device_ds" {
  device_name = "testacc_Device_ds"
  host        = "testacc-ds.device"

  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDeviceConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_Device_ds" {
  device_name = "testacc_Device_ds"
  host        = "testacc-ds.device"

  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
}

data "wallix-bastion_device" "testacc_dataDevice" {
  device_name = wallix-bastion_device.testacc_Device_ds.device_name
}
`
}
