package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceApplicationLocalDomainAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceApplicationLocalDomainAccountConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceApplicationLocalDomainAccountConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccount",
						"account_name", "testacc_dataAppLocalDomAccount_ds"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccount",
						"account_login", "testacc_dataAppLocalDomAccount_ds"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccount",
						"application_id",
						"wallix-bastion_application.testacc_dataAppLocalDomAccount_ds", "id"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccount",
						"domain_id",
						"wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccount_ds", "id"),
				),
			},
		},
	})
}

// Resource creation configuration.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainAccountConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDomAccount_ds" {
  device_name = "testacc_dataAppLocalDomAccount_ds"
  host        = "192.168.100.15"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDomAccount_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.id
  service_name      = "testacc_dataAppLocalDomAccount_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDomAccount_ds" {
  cluster_name = "testacc_dataAppLocalDomAccount_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccount_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDomAccount_ds" {
  application_name  = "testacc_dataAppLocalDomAccount_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccount_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDomAccount_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDomAccount_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccount_ds.id
  domain_name    = "testacc_dataAppLocalDomAccount_ds"
}

resource "wallix-bastion_application_localdomain_account" "testacc_dataAppLocalDomAccount_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccount_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccount_ds.id
  account_name   = "testacc_dataAppLocalDomAccount_ds"
  account_login  = "testacc_dataAppLocalDomAccount_ds"
}
`
}

// Datasource configuration to retrieve the created resource.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainAccountConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDomAccount_ds" {
  device_name = "testacc_dataAppLocalDomAccount_ds"
  host        = "192.168.100.15"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDomAccount_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.id
  service_name      = "testacc_dataAppLocalDomAccount_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDomAccount_ds" {
  cluster_name = "testacc_dataAppLocalDomAccount_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccount_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDomAccount_ds" {
  application_name  = "testacc_dataAppLocalDomAccount_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDomAccount_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccount_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDomAccount_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDomAccount_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccount_ds.id
  domain_name    = "testacc_dataAppLocalDomAccount_ds"
}

resource "wallix-bastion_application_localdomain_account" "testacc_dataAppLocalDomAccount_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccount_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccount_ds.id
  account_name   = "testacc_dataAppLocalDomAccount_ds"
  account_login  = "testacc_dataAppLocalDomAccount_ds"
}

data "wallix-bastion_application_localdomain_account" "testacc_dataAppLocalDomAccount" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccount_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccount_ds.id
  account_name   = wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccount_ds.account_name
}
`
}
