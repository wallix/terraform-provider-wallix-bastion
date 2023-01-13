package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDeviceLocalDomainAccountCred_basic(t *testing.T) {
	resourceName := "wallix-bastion_device_localdomain_account_credential.testacc_DeviceLocalDomainAccountCred2"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDeviceLocalDomainAccountCredCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceDeviceLocalDomainAccountCredUpdate(),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Resource %s not found", resourceName)
					}
					devID := rs.Primary.Attributes["device_id"]
					if devID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "device_id", rs.Primary.Attributes)
					}
					domID := rs.Primary.Attributes["domain_id"]
					if domID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "domain_id", rs.Primary.Attributes)
					}
					accID := rs.Primary.Attributes["account_id"]
					if accID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "account_id", rs.Primary.Attributes)
					}

					return devID + "/" + domID + "/" + accID + "/ssh_key", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDeviceLocalDomainAccountCredCreate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccountCred" {
  device_name = "testacc_DeviceLocalDomainAccountCred"
  host        = "testacc_localdomain_account.device"
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

func testAccResourceDeviceLocalDomainAccountCredUpdate() string {
	return `
resource "wallix-bastion_device" "testacc_DeviceLocalDomainAccountCred" {
  device_name = "testacc_DeviceLocalDomainAccountCred"
  host        = "testacc_localdomain_account.device"
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
resource "wallix-bastion_device_localdomain_account_credential" "testacc_DeviceLocalDomainAccountCred2" {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccountCred.id
  domain_id   = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccountCred.id
  account_id  = wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccountCred.id
  type        = "ssh_key"
  private_key = tls_private_key.testacc_DomainAccountCred.private_key_pem
  passphrase  = random_password.testacc_DomainAccountCred.result
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "tls_private_key" "testacc_DomainAccountCred" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
`
}
