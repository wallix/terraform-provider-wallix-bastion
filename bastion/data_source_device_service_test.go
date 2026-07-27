package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDeviceService_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDeviceServiceConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDeviceServiceConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"service_name", "testacc_DeviceService"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"connection_policy", "SSH"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"port", "22"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"protocol", "SSH"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"subprotocols.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"subprotocols.*", "SSH_SHELL_SESSION"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"global_domains.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_device_service.testacc_dataDeviceService",
						"global_domains.*", "testacc_DeviceService"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDeviceServiceConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceService" {
  device_name = "testacc_DeviceService"
  host        = "192.168.100.2"
}
resource "wallix-bastion_domain" "testacc_DeviceService" {
  domain_name = "testacc_DeviceService"
}
resource "wallix-bastion_device_service" "testacc_DeviceService" {
  device_id         = wallix-bastion_device.testacc_DeviceService.id
  service_name      = "testacc_DeviceService"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
  global_domains    = [wallix-bastion_domain.testacc_DeviceService.domain_name]
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDeviceServiceConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceService" {
  device_name = "testacc_DeviceService"
  host        = "192.168.100.2"
}
resource "wallix-bastion_domain" "testacc_DeviceService" {
  domain_name = "testacc_DeviceService"
}
resource "wallix-bastion_device_service" "testacc_DeviceService" {
  device_id         = wallix-bastion_device.testacc_DeviceService.id
  service_name      = "testacc_DeviceService"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
  global_domains    = [wallix-bastion_domain.testacc_DeviceService.domain_name]
}

data "wallix-bastion_device_service" "testacc_dataDeviceService" {
  device_id    = wallix-bastion_device.testacc_DeviceService.id
  service_name = wallix-bastion_device_service.testacc_DeviceService.service_name
}
`
}
