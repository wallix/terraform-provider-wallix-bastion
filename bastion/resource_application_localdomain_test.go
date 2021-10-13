package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceApplicationLocalDomain_basic(t *testing.T) {
	resourceName := "wallix-bastion_application_localdomain.testacc_AppLocalDom"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationLocalDomainCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceApplicationLocalDomainUpdate(),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Resource %s not found", resourceName)
					}
					appID := rs.Primary.Attributes["application_id"]
					if appID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "application_id", rs.Primary.Attributes)
					}

					return appID + "/testacc_AppLocalDom", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll
func testAccResourceApplicationLocalDomainCreate() string {
	return `
resource wallix-bastion_device testacc_AppLocalDom {
  device_name = "testacc_AppLocalDom"
  host        = "testacc_AppLocalDom"
}

resource wallix-bastion_device_service testacc_AppLocalDom {
  device_id         = wallix-bastion_device.testacc_AppLocalDom.id
  service_name      = "testacc_AppLocalDom"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_AppLocalDom {
  cluster_name = "testacc_AppLocalDom"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDom.device_name}:${wallix-bastion_device_service.testacc_AppLocalDom.service_name}",
  ]
}

resource wallix-bastion_application testacc_AppLocalDom {
  application_name  = "testacc_AppLocalDom_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDom.device_name}:${wallix-bastion_device_service.testacc_AppLocalDom.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDom.cluster_name
}

resource wallix-bastion_application_localdomain_account testacc_AppLocalDom {
  application_id = wallix-bastion_application.testacc_AppLocalDom.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDom.id
  account_name   = "testacc_AppLocalDom"
  account_login  = "testacc_AppLocalDom"
}

resource wallix-bastion_application_localdomain testacc_AppLocalDom {
  application_id = wallix-bastion_application.testacc_AppLocalDom.id
  domain_name    = "testacc_AppLocalDom"
}
`
}

// nolint: lll
func testAccResourceApplicationLocalDomainUpdate() string {
	return `
resource wallix-bastion_device testacc_AppLocalDom {
  device_name = "testacc_AppLocalDom"
  host        = "testacc_AppLocalDom"
}

resource wallix-bastion_device_service testacc_AppLocalDom {
  device_id         = wallix-bastion_device.testacc_AppLocalDom.id
  service_name      = "testacc_AppLocalDom"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_AppLocalDom {
  cluster_name = "testacc_AppLocalDom"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDom.device_name}:${wallix-bastion_device_service.testacc_AppLocalDom.service_name}",
  ]
}

resource wallix-bastion_application testacc_AppLocalDom {
  application_name  = "testacc_AppLocalDom_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDom.device_name}:${wallix-bastion_device_service.testacc_AppLocalDom.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDom.cluster_name
}

resource wallix-bastion_application_localdomain_account testacc_AppLocalDom {
  application_id = wallix-bastion_application.testacc_AppLocalDom.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDom.id
  account_name   = "testacc_AppLocalDom"
  account_login  = "testacc_AppLocalDom"
}

resource wallix-bastion_application_localdomain testacc_AppLocalDom {
  application_id         = wallix-bastion_application.testacc_AppLocalDom.id
  domain_name            = "testacc_AppLocalDom"
  description            = "test"
  admin_account          = "testacc_AppLocalDom"
  enable_password_change = true
  password_change_policy = "default"
  password_change_plugin = "MySQL"
  password_change_plugin_parameters = jsonencode({
	"host": "10.11.12.13",
	"port": 3306
  })
}
`
}
