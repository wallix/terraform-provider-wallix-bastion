package bastion_test

import (
	"os"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"
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

func TestAccResourceApplication_jumphost(t *testing.T) {
	if os.Getenv("TESTACC_JUMPHOST") != "" {
		if v := os.Getenv("WALLIX_BASTION_API_VERSION"); semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testAccResourceApplicationCreateJumphost(),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(
								"wallix-bastion_application.testacc_Appli",
								"id"),
						),
					},
					{
						Config: testAccResourceApplicationUpdateJumphost(),
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
	}
}

// nolint: lll, nolintlint
func testAccResourceApplicationCreate() string {
	return `
resource "wallix-bastion_device" "testacc_App" {
  device_name = "testacc_App"
  host        = "testacc_App"
}

resource "wallix-bastion_device_service" "testacc_App" {
  device_id         = wallix-bastion_device.testacc_App.id
  service_name      = "testacc_App"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_App" {
  cluster_name = "testacc_App"
  interactive_logins = [
    "${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_Appli" {
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

// nolint: lll, nolintlint
func testAccResourceApplicationCreateJumphost() string {
	return `
resource "wallix-bastion_application" "testacc_Appli" {
  application_name  = "testacc_Appli_jumphost"
  category          = "jumphost"
  connection_policy = "JumpHost"
  browser           = "Mozilla Firefox"
  browser_version   = "93.0"
  application_url   = "https://github.com/login"
}
`
}

// nolint: lll, nolintlint
func testAccResourceApplicationUpdate() string {
	return `
resource "wallix-bastion_device" "testacc_App" {
  device_name = "testacc_App"
  host        = "testacc_App"
}

resource "wallix-bastion_device_service" "testacc_App" {
  device_id         = wallix-bastion_device.testacc_App.id
  service_name      = "testacc_App"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_App" {
  cluster_name = "testacc_App"
  interactive_logins = [
    "${wallix-bastion_device.testacc_App.device_name}:${wallix-bastion_device_service.testacc_App.service_name}",
  ]
}

resource "wallix-bastion_domain" "testacc_App" {
  domain_name = "testacc_App"
}

resource "wallix-bastion_application" "testacc_Appli" {
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

func testAccResourceApplicationUpdateJumphost() string {
	return `
resource "wallix-bastion_application" "testacc_Appli" {
  application_name  = "testacc_Appli_jumphost"
  description       = "testacc Appli jumphost"
  category          = "jumphost"
  connection_policy = "JumpHost"
  browser           = "Google Chrome"
  browser_version   = "94.0.4606.81-1"
  application_url   = "https://github.com/login"
  parameters        = "app_parameters"
}
`
}
