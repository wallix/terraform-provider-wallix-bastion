package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDevice_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDeviceCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_device.testacc_Device",
						"id"),

					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"wallix-bastion_device.testacc_Device",
						"tags.*", map[string]string{
							"key":   "testkey",
							"value": "testvalue",
						}),
				),
			},
			{
				Config: testAccResourceDeviceUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"alias", "testacc-Device"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"description", "testacc Device"),

					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"wallix-bastion_device.testacc_Device",
						"tags.*", map[string]string{
							"key":   "testkey",
							"value": "testvalue",
						}),

					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"local_domains.#", "1"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_device.testacc_Device",
						"services.#", "1"),
				),
			},
			{
				ResourceName:  "wallix-bastion_device.testacc_Device",
				ImportState:   true,
				ImportStateId: "testacc_Device",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDeviceCreate() string {
	return `
resource "wallix-bastion_device" "testacc_Device" {
  device_name = "testacc_Device"
  host        = "testacc.device"
  
  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
}

resource "wallix-bastion_device_localdomain" "testacc_Device" {
  device_id   = wallix-bastion_device.testacc_Device.id
  domain_name = "testacc_Device"
}
resource "wallix-bastion_device_service" "testacc_Device" {
  device_id         = wallix-bastion_device.testacc_Device.id
  service_name      = "testacc_Device"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
}
`
}

func testAccResourceDeviceUpdate() string {
	return `
resource "wallix-bastion_device" "testacc_Device" {
  device_name = "testacc_Device"
  host        = "testacc.device"
  alias       = "testacc-Device"
  description = "testacc Device"
  
  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }	
}
resource "wallix-bastion_device_localdomain" "testacc_Device" {
  device_id   = wallix-bastion_device.testacc_Device.id
  domain_name = "testacc_Device"
}
resource "wallix-bastion_device_service" "testacc_Device" {
  device_id         = wallix-bastion_device.testacc_Device.id
  service_name      = "testacc_Device"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
}
`
}
