package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDeviceLocalDomainAccountCred_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			tvRandomProviderName: {
				Source: tvRandomProviderSource,
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDeviceLocalDomainAccountCredConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDeviceLocalDomainAccountCredConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.wallix-bastion_device_localdomain_account_credential.testacc_dataDeviceLocalDomainAccountCred",
						"id"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_device_localdomain_account_credential.testacc_dataDeviceLocalDomainAccountCred",
						"type", "ssh_key"),
					resource.TestCheckResourceAttrSet(
						"data.wallix-bastion_device_localdomain_account_credential.testacc_dataDeviceLocalDomainAccountCred",
						"public_key"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDeviceLocalDomainAccountCredConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccountCred" {
  device_name = "testacc_DeviceLocalDomainAccountCred"
  host        = "192.168.100.4"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomainAccountCred" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_name = "testacc_DeviceLocalDomainAccountCred"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomainAccountCred" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_name  = "testacc_DeviceLocalDomainAccountCred_admin"
  account_login = "admin"
}
resource "wallix-bastion_device_localdomain_account_credential" "testacc_DeviceLocalDomainAccountCred" {
  device_id  = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id  = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type       = "password"
  password   = random_password.testacc_DomainAccountCred.result
}
resource "wallix-bastion_device_localdomain_account_credential" "testacc_DeviceLocalDomainAccountCred2" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id   = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id  = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type        = "ssh_key"
  private_key = "generate:RSA_4096"
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDeviceLocalDomainAccountCredConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccountCred" {
  device_name = "testacc_DeviceLocalDomainAccountCred"
  host        = "192.168.100.4"
}
resource "wallix-bastion_device_localdomain" "testacc_DeviceLocalDomainAccountCred" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_name = "testacc_DeviceLocalDomainAccountCred"
}
resource "wallix-bastion_device_localdomain_account" "testacc_DeviceLocalDomainAccountCred" {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_name  = "testacc_DeviceLocalDomainAccountCred_admin"
  account_login = "admin"
}
resource "wallix-bastion_device_localdomain_account_credential" "testacc_DeviceLocalDomainAccountCred" {
  device_id  = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id  = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type       = "password"
  password   = random_password.testacc_DomainAccountCred.result
}
resource "wallix-bastion_device_localdomain_account_credential" "testacc_DeviceLocalDomainAccountCred2" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id   = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id  = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type        = "ssh_key"
  private_key = "generate:RSA_4096"
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}

data "wallix-bastion_device_localdomain_account_credential" "testacc_dataDeviceLocalDomainAccountCred" {
  device_id  = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id  = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type       = wallix-bastion_device_localdomain_account_credential.testacc_DeviceLocalDomainAccountCred2.type
}
`
}
