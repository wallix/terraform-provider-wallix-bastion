package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDeviceService_basic(t *testing.T) {
	resourceName := "wallix-bastion_device_service.testacc_DeviceService"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDeviceServiceCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceDeviceServiceUpdate(),
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

					return devID + "/testacc_DeviceService", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDeviceServiceCreate() string {
	return `
resource wallix-bastion_device testacc_DeviceService {
  device_name = "testacc_DeviceService"
  host        = "testacc_service.device"
}
resource wallix-bastion_domain testacc_DeviceService {
  domain_name = "testacc_DeviceService"
}
resource wallix-bastion_device_service testacc_DeviceService {
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

func testAccResourceDeviceServiceUpdate() string {
	return `
resource wallix-bastion_device testacc_DeviceService {
  device_name = "testacc_DeviceService"
  host        = "testacc_service.device"
}
resource wallix-bastion_domain testacc_DeviceService {
  domain_name = "testacc_DeviceService"
}
resource wallix-bastion_device_service testacc_DeviceService {
  device_id         = wallix-bastion_device.testacc_DeviceService.id
  service_name      = "testacc_DeviceService"
  connection_policy = "SSH"
  port              = 2242
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
  global_domains    = [wallix-bastion_domain.testacc_DeviceService.domain_name]
}
`
}
