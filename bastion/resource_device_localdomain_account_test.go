package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDeviceLocalDomainAccount_basic(t *testing.T) {
	resourceName := "wallix-bastion_device_localdomain_account.testacc_DeviceLocalDomainAccount"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDeviceLocalDomainAccountCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceDeviceLocalDomainAccountUpdate(),
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

					return devID + "/" + domID + "/testacc_DeviceLocalDomainAccount_admin", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDeviceLocalDomainAccountCreate() string {
	return `
resource wallix-bastion_device testacc_DeviceLocalDomainAccount {
  device_name = "testacc_DeviceLocalDomainAccount"
  host        = "testacc_localdomain_account.device"
}
resource wallix-bastion_device_localdomain testacc_DeviceLocalDomainAccount {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_name = "testacc_DeviceLocalDomainAccount"
}
resource wallix-bastion_device_localdomain_account testacc_DeviceLocalDomainAccount {
  device_id     = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_id     = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccount.id
  account_name  = "testacc_DeviceLocalDomainAccount_admin"
  account_login = "admin"
}
`
}

func testAccResourceDeviceLocalDomainAccountUpdate() string {
	return `
resource wallix-bastion_device testacc_DeviceLocalDomainAccount {
  device_name = "testacc_DeviceLocalDomainAccount"
  host        = "testacc_localdomain_account.device"
}
resource wallix-bastion_device_localdomain testacc_DeviceLocalDomainAccount {
  device_id   = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_name = "testacc_DeviceLocalDomainAccount"
}
resource wallix-bastion_device_service testacc_DeviceLocalDomainAccount {
  device_id         = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  service_name      = "testacc_DeviceLocalDomainAccount"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
}
resource wallix-bastion_device_localdomain_account testacc_DeviceLocalDomainAccount {
  device_id            = wallix-bastion_device.testacc_DeviceLocalDomainAccount.id
  domain_id            = wallix-bastion_device_localdomain.testacc_DeviceLocalDomainAccount.id
  account_name         = "testacc_DeviceLocalDomainAccount_admin"
  account_login        = "admin"
  auto_change_password = true
  auto_change_ssh_key  = true
  certificate_validity = "+2h"
  description          = "testacc DeviceLocalDomainAccount"
  services             = [wallix-bastion_device_service.testacc_DeviceLocalDomainAccount.service_name]
}
`
}
