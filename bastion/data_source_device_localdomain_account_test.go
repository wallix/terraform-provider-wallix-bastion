package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDeviceLocalDomainAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDeviceLocalDomainAccountConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDeviceLocalDomainAccountConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.wallix-bastion_device_localdomain_account.testacc_dataDeviceLocalDomainAccount",
						"id"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_device_localdomain_account.testacc_dataDeviceLocalDomainAccount",
						"account_name", "testacc_DeviceLocalDomainAccount_admin"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_device_localdomain_account.testacc_dataDeviceLocalDomainAccount",
						"account_login", "admin"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDeviceLocalDomainAccountConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccount" {
  device_name = "testacc_DeviceLocalDomainAccount"
  host        = "192.168.100.4"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomainAccount" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_name = "testacc_DeviceLocalDomainAccount"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomainAccount" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccount.id
  account_name  = "testacc_DeviceLocalDomainAccount_admin"
  account_login = "admin"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDeviceLocalDomainAccountConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccount" {
  device_name = "testacc_DeviceLocalDomainAccount"
  host        = "192.168.100.4"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomainAccount" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_name = "testacc_DeviceLocalDomainAccount"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomainAccount" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccount.id
  account_name  = "testacc_DeviceLocalDomainAccount_admin"
  account_login = "admin"
}

data "wallix-bastion_device_localdomain_account" "testacc_dataDeviceLocalDomainAccount" {
  device_id    = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_id    = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccount.id
  account_name = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccount.account_name
}
`
}
