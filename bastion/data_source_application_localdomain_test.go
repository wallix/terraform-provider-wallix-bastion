package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceApplicationLocalDomain_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceApplicationLocalDomainConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceApplicationLocalDomainConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_application_localdomain.testacc_dataAppLocalDom",
						"domain_name", "testacc_dataAppLocalDom_ds"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain.testacc_dataAppLocalDom", "application_id",
						"wallix-bastion_application.testacc_dataAppLocalDom_ds", "id"),
				),
			},
		},
	})
}

// Resource creation configuration.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDom_ds" {
  device_name = "testacc_dataAppLocalDom_ds"
  host        = "192.168.100.14"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDom_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDom_ds.id
  service_name      = "testacc_dataAppLocalDom_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDom_ds" {
  cluster_name = "testacc_dataAppLocalDom_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDom_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDom_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDom_ds" {
  application_name  = "testacc_dataAppLocalDom_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDom_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDom_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDom_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDom_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDom_ds.id
  domain_name    = "testacc_dataAppLocalDom_ds"
}
`
}

// Datasource configuration to retrieve the created resource.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDom_ds" {
  device_name = "testacc_dataAppLocalDom_ds"
  host        = "192.168.100.14"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDom_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDom_ds.id
  service_name      = "testacc_dataAppLocalDom_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDom_ds" {
  cluster_name = "testacc_dataAppLocalDom_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDom_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDom_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDom_ds" {
  application_name  = "testacc_dataAppLocalDom_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDom_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDom_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDom_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDom_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDom_ds.id
  domain_name    = "testacc_dataAppLocalDom_ds"
}

data "wallix-bastion_application_localdomain" "testacc_dataAppLocalDom" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDom_ds.id
  domain_name    = wallix-bastion_application_localdomain.testacc_dataAppLocalDom_ds.domain_name
}
`
}
