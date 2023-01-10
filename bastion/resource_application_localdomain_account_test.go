package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceApplicationLocalDomainAccount_basic(t *testing.T) {
	resourceName := "wallix-bastion_application_localdomain_account.testacc_AppLocalDomAccount"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationLocalDomainAccountCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_application_localdomain_account.testacc_AppLocalDomAccount",
						"id"),
				),
			},
			{
				Config: testAccResourceApplicationLocalDomainAccountUpdate(),
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
					domID := rs.Primary.Attributes["domain_id"]
					if domID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "domain_id", rs.Primary.Attributes)
					}

					return appID + "/" + domID + "/testacc_AppLocalDomAccount", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll, nolintlint
func testAccResourceApplicationLocalDomainAccountCreate() string {
	return `
resource wallix-bastion_device testacc_AppLocalDomAccount {
  device_name = "testacc_AppLocalDomAccount"
  host        = "testacc_AppLocalDomAccount"
}

resource wallix-bastion_device_service testacc_AppLocalDomAccount {
  device_id         = wallix-bastion_device.testacc_AppLocalDomAccount.id
  service_name      = "testacc_AppLocalDomAccount"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_AppLocalDomAccount {
  cluster_name = "testacc_AppLocalDomAccount"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDomAccount.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccount.service_name}",
  ]
}

resource wallix-bastion_application testacc_AppLocalDomAccount {
  application_name  = "testacc_AppLocalDomAccount_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDomAccount.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccount.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDomAccount.cluster_name
}

resource wallix-bastion_application_localdomain testacc_AppLocalDomAccount {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccount.id
  domain_name    = "testacc_AppLocalDomAccount"
}

resource wallix-bastion_application_localdomain_account testacc_AppLocalDomAccount {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccount.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDomAccount.id
  account_name   = "testacc_AppLocalDomAccount"
  account_login  = "testacc_AppLocalDomAccount"
}
`
}

// nolint: lll, nolintlint
func testAccResourceApplicationLocalDomainAccountUpdate() string {
	return `
resource wallix-bastion_device testacc_AppLocalDomAccount {
  device_name = "testacc_AppLocalDomAccount"
  host        = "testacc_AppLocalDomAccount"
}

resource wallix-bastion_device_service testacc_AppLocalDomAccount {
  device_id         = wallix-bastion_device.testacc_AppLocalDomAccount.id
  service_name      = "testacc_AppLocalDomAccount"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_AppLocalDomAccount {
  cluster_name = "testacc_AppLocalDomAccount"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDomAccount.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccount.service_name}",
  ]
}

resource wallix-bastion_application testacc_AppLocalDomAccount {
  application_name  = "testacc_AppLocalDomAccount_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDomAccount.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccount.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDomAccount.cluster_name
}

resource wallix-bastion_application_localdomain testacc_AppLocalDomAccount {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccount.id
  domain_name    = "testacc_AppLocalDomAccount"
}

resource wallix-bastion_application_localdomain_account testacc_AppLocalDomAccount {
  application_id       = wallix-bastion_application.testacc_AppLocalDomAccount.id
  domain_id            = wallix-bastion_application_localdomain.testacc_AppLocalDomAccount.id
  account_name         = "testacc_AppLocalDomAccount"
  account_login        = "testacc_AppLocalDomAccount"
  auto_change_password = true
  description          = "test"
  password             = "password"
}
`
}
