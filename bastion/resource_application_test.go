package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceApplication_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_application.testacc_Appli",
						"id"),
				),
			},
			{
				Config: testAccResourceApplicationUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_application.testacc_Appli",
				ImportState:   true,
				ImportStateId: "testacc_Appli",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll
func testAccResourceApplicationCreate() string {
	return `
resource wallix-bastion_device testacc_App {
  device_name = "testacc_App"
  host        = "testacc_App"
}

resource wallix-bastion_device_service testacc_App {
  device_id         = wallix-bastion_device.testacc_App.id
  service_name      = "testacc_App"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_App {
  cluster_name = "testacc_App"
  interactive_logins = [
    "${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}",
  ]
}

resource wallix-bastion_application testacc_Appli {
  application_name  = "testacc_Appli"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_App.cluster_name
}
`
}

// nolint: lll
func testAccResourceApplicationUpdate() string {
	return `
resource wallix-bastion_device testacc_App {
  device_name = "testacc_App"
  host        = "testacc_App"
}

resource wallix-bastion_device_service testacc_App {
  device_id         = wallix-bastion_device.testacc_App.id
  service_name      = "testacc_App"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource wallix-bastion_cluster testacc_App {
  cluster_name = "testacc_App"
  interactive_logins = [
    "${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}",
  ]
}

resource wallix-bastion_domain testacc_App {
  domain_name = "testacc_App"
}

resource wallix-bastion_application testacc_Appli {
  application_name  = "testacc_Appli"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target         = wallix-bastion_cluster.testacc_App.cluster_name
  parameters     = "app_parameters"
  global_domains = [wallix-bastion_domain.testacc_App.domain_name]
}
`
}
