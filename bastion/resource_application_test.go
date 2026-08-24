package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func TestAccResourceApplication_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_application.testacc_Appli",
						"id"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_application.testacc_Appli",
						"tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"wallix-bastion_application.testacc_Appli",
						"tags.*", map[string]string{
							"key":   "testkey",
							"value": "testvalue",
						}),
				),
			},
			{
				Config: testAccResourceApplicationUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_application.testacc_Appli",
						"parameters", "app_parameters"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_application.testacc_Appli",
						"global_domains.#", "1"),
				),
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

func TestAccResourceApplication_web(t *testing.T) {
	if os.Getenv("TESTACC_WEB_APP") != "" {
		if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccResourceApplicationCreateWeb(),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(
								"wallix-bastion_application.testacc_Appli_web",
								"id"),
							resource.TestCheckResourceAttr(
								"wallix-bastion_application.testacc_Appli_web",
								"category", "web_application"),
							resource.TestCheckResourceAttr(
								"wallix-bastion_application.testacc_Appli_web",
								"application_url", "https://github.com/login"),
						),
					},
					{
						Config: testAccResourceApplicationUpdateWeb(),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"wallix-bastion_application.testacc_Appli_web",
								"description", "testacc Web Application"),
							resource.TestCheckResourceAttr(
								"wallix-bastion_application.testacc_Appli_web",
								"application_url", "https://github.com/login"),
						),
					},
					{
						ResourceName:  "wallix-bastion_application.testacc_Appli_web",
						ImportState:   true,
						ImportStateId: "testacc_Appli_web",
					},
				},
				PreventPostDestroyRefresh: true,
			})
		}
	}
}

// Deprecated: jumphost category is no longer supported in API v3.12+.
// This test is kept for backward compatibility testing with older API versions.
func TestAccResourceApplication_jumphost_deprecated(t *testing.T) {
	if os.Getenv("TESTACC_JUMPHOST") != "" {
		v := os.Getenv("WALLIX_BASTION_API_VERSION")
		// Only run this test for API versions < 3.12 where jumphost is still supported
		if v != "" && semver.Compare(v, bastion.VersionWallixAPI312) < 0 {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccResourceApplicationCreateJumphost(),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(
								"wallix-bastion_application.testacc_Appli_jumphost",
								"id"),
							resource.TestCheckResourceAttr(
								"wallix-bastion_application.testacc_Appli_jumphost",
								"category", "jumphost"),
						),
					},
					{
						Config: testAccResourceApplicationUpdateJumphost(),
					},
					{
						ResourceName:  "wallix-bastion_application.testacc_Appli_jumphost",
						ImportState:   true,
						ImportStateId: "testacc_Appli_jumphost",
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
  host        = "192.168.100.11"
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
  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
}
`
}

// nolint: lll, nolintlint
func testAccResourceApplicationCreateJumphost() string {
	return `
resource "wallix-bastion_application" "testacc_Appli_jumphost" {
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
func testAccResourceApplicationCreateWeb() string {
	return `
resource "wallix-bastion_application" "testacc_Appli_web" {
  application_name  = "testacc_Appli_web"
  category          = "web_application"
  connection_policy = "WebApp"
  application_url   = "https://github.com/login"
}
`
}

// nolint: lll, nolintlint
func testAccResourceApplicationUpdate() string {
	return `
resource "wallix-bastion_device" "testacc_App" {
  device_name = "testacc_App"
  host        = "192.168.100.11"
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
  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
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

func testAccResourceApplicationUpdateWeb() string {
	return `
resource "wallix-bastion_application" "testacc_Appli_web" {
  application_name  = "testacc_Appli_web"
  description       = "testacc Web Application"
  category          = "web_application"
  connection_policy = "WebApp"
  application_url   = "https://github.com/login"
  parameters        = "app_parameters"
}
`
}
