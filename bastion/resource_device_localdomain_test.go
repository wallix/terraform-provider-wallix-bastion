package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDeviceLocalDomain_basic(t *testing.T) {
	resourceName := "wallix-bastion_device_localdomain.testacc_DeviceLocalDomain"
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
				Config: testAccResourceDeviceLocalDomainCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceDeviceLocalDomainUpdate(),
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

					return devID + "/testacc_DeviceLocalDomain", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDeviceLocalDomainCreate() string {
	return `
resource wallix-bastion_device testacc_DeviceLocalDomain {
  device_name = "testacc_DeviceLocalDomain"
  host        = "testacc_localdomain.device"
}
resource wallix-bastion_device_localdomain testacc_DeviceLocalDomain {
  device_id      = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_name    = "testacc_DeviceLocalDomain"
  ca_private_key = "generate:RSA_4096"
}
resource wallix-bastion_device_localdomain_account testacc_DeviceLocalDomain {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomain.id
  account_name  = "testacc_DeviceLocalDomain_admin"
  account_login = "admin"
}
`
}

func testAccResourceDeviceLocalDomainUpdate() string {
	return `
resource wallix-bastion_device testacc_DeviceLocalDomain {
  device_name = "testacc_DeviceLocalDomain"
  host        = "testacc_localdomain.device"
}
resource wallix-bastion_device_localdomain testacc_DeviceLocalDomain {
  device_id                         = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_name                       = "testacc_DeviceLocalDomain"
  description                       = "testacc DeviceLocalDomain"
  ca_private_key                    = tls_private_key.testacc_DeviceLocalDomain.private_key_pem
  passphrase                        = random_password.testacc_DeviceLocalDomain.result
  admin_account                     = "testacc_DeviceLocalDomain_admin"
  enable_password_change            = true
  password_change_policy            = "default"
  password_change_plugin            = "Unix"
  password_change_plugin_parameters = jsonencode({})
}
resource wallix-bastion_device_localdomain_account testacc_DeviceLocalDomain {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomain.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomain.id
  account_name  = "testacc_DeviceLocalDomain_admin"
  account_login = "admin"
}
resource "tls_private_key" "testacc_DeviceLocalDomain" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
resource "random_password" "testacc_DeviceLocalDomain" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
`
}
