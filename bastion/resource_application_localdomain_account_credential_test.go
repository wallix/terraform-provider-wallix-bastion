package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceApplicationLocalDomainAccountCred_basic(t *testing.T) {
	resourceName := "wallix-bastion_application_localdomain_account_credential.testacc_AppLocalDomAccountCred"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationLocalDomainAccountCredCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"type", "password"),
				),
			},
			{
				Config: testAccResourceApplicationLocalDomainAccountCredUpdate(),
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
					accID := rs.Primary.Attributes["account_id"]
					if accID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "account_id", rs.Primary.Attributes)
					}

					return appID + "/" + domID + "/" + accID + "/password", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll, nolintlint
func testAccResourceApplicationLocalDomainAccountCredCreate() string {
	return `
resource "wallix-bastion_device" "testacc_AppLocalDomAccountCred" {
  device_name = "testacc_AppLocalDomAccountCred"
  host        = "192.168.100.12"
}

resource "wallix-bastion_device_service" "testacc_AppLocalDomAccountCred" {
  device_id         = wallix-bastion_device.testacc_AppLocalDomAccountCred.id
  service_name      = "testacc_AppLocalDomAccountCred"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_AppLocalDomAccountCred" {
  cluster_name = "testacc_AppLocalDomAccountCred"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDomAccountCred.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccountCred.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_AppLocalDomAccountCred" {
  application_name  = "testacc_AppLocalDomAccountCred_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDomAccountCred.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccountCred.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDomAccountCred.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_name    = "testacc_AppLocalDomAccountCred"
}

resource "wallix-bastion_application_localdomain_account" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDomAccountCred.id
  account_name   = "testacc_AppLocalDomAccountCred_admin"
  account_login  = "admin"
}

resource "wallix-bastion_application_localdomain_account_credential" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDomAccountCred.id
  account_id     = wallix-bastion_application_localdomain_account.testacc_AppLocalDomAccountCred.id
  type           = "password"
  password       = random_password.testacc_AppLocalDomAccountCred.result
}
resource "random_password" "testacc_AppLocalDomAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
`
}

// nolint: lll, nolintlint
func testAccResourceApplicationLocalDomainAccountCredUpdate() string {
	return `
resource "wallix-bastion_device" "testacc_AppLocalDomAccountCred" {
  device_name = "testacc_AppLocalDomAccountCred"
  host        = "192.168.100.12"
}

resource "wallix-bastion_device_service" "testacc_AppLocalDomAccountCred" {
  device_id         = wallix-bastion_device.testacc_AppLocalDomAccountCred.id
  service_name      = "testacc_AppLocalDomAccountCred"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_AppLocalDomAccountCred" {
  cluster_name = "testacc_AppLocalDomAccountCred"
  interactive_logins = [
    "${wallix-bastion_device.testacc_AppLocalDomAccountCred.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccountCred.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_AppLocalDomAccountCred" {
  application_name  = "testacc_AppLocalDomAccountCred_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_AppLocalDomAccountCred.device_name}:${wallix-bastion_device_service.testacc_AppLocalDomAccountCred.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_AppLocalDomAccountCred.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_name    = "testacc_AppLocalDomAccountCred"
}

resource "wallix-bastion_application_localdomain_account" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDomAccountCred.id
  account_name   = "testacc_AppLocalDomAccountCred_admin"
  account_login  = "admin"
}

resource "wallix-bastion_application_localdomain_account_credential" "testacc_AppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_AppLocalDomAccountCred.id
  domain_id      = wallix-bastion_application_localdomain.testacc_AppLocalDomAccountCred.id
  account_id     = wallix-bastion_application_localdomain_account.testacc_AppLocalDomAccountCred.id
  type           = "password"
  password       = random_password.testacc_AppLocalDomAccountCred2.result
}
resource "random_password" "testacc_AppLocalDomAccountCred2" {
  length           = 14
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
`
}
