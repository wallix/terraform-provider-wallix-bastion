package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDeviceLocalDomain_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDeviceLocalDomainConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDeviceLocalDomainConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_device_localdomain.testacc_dataDeviceLocalDomain",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_device_localdomain.testacc_dataDeviceLocalDomain",
						"domain_name", "testacc_DeviceLocalDomain"),
					resource.TestCheckResourceAttrSet("data.wallix-bastion_device_localdomain.testacc_dataDeviceLocalDomain",
						"ca_public_key"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDeviceLocalDomainConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomain" {
  device_name = "testacc_DeviceLocalDomain"
  host        = "192.168.100.3"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomain" {
  device_id      = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_name    = "testacc_DeviceLocalDomain"
  ca_private_key = "generate:RSA_4096"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomain" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomain.id
  account_name  = "testacc_DeviceLocalDomain_admin"
  account_login = "admin"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDeviceLocalDomainConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomain" {
  device_name = "testacc_DeviceLocalDomain"
  host        = "192.168.100.3"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomain" {
  device_id      = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_name    = "testacc_DeviceLocalDomain"
  ca_private_key = "generate:RSA_4096"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomain" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomain.id
  account_name  = "testacc_DeviceLocalDomain_admin"
  account_login = "admin"
}

data "wallix-bastion_device_localdomain" "testacc_dataDeviceLocalDomain" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_name = wallix-bastion_device_localdomain.testacc_DeviceLocalDomain.domain_name
}
`
}
