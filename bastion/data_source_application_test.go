package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceApplication_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceApplicationConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceApplicationConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_application.testacc_dataAppli",
						"application_name", "testacc_dataAppli_ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_application.testacc_dataAppli",
						"connection_policy", "RDP"),
					resource.TestCheckResourceAttr("data.wallix-bastion_application.testacc_dataAppli",
						"tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_application.testacc_dataAppli",
						"tags.*", map[string]string{
							tvKey:   tvTestKey,
							tvValue: tvTestValue,
						}),
					resource.TestCheckResourceAttr("data.wallix-bastion_application.testacc_dataAppli",
						"paths.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_application.testacc_dataAppli",
						"paths.*", map[string]string{
							"program":     "application_path",
							"working_dir": "directory",
						}),
				),
			},
		},
	})
}

// Resource creation configuration.
// nolint: lll, nolintlint
func testAccDataSourceApplicationConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppli_ds" {
  device_name = "testacc_dataAppli_ds"
  host        = "192.168.100.13"
}

resource "wallix-bastion_device_service" "testacc_dataAppli_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppli_ds.id
  service_name      = "testacc_dataAppli_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppli_ds" {
  cluster_name = "testacc_dataAppli_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppli_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppli_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppli_ds" {
  application_name  = "testacc_dataAppli_ds"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppli_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppli_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppli_ds.cluster_name
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

// Datasource configuration to retrieve the created resource.
// nolint: lll, nolintlint
func testAccDataSourceApplicationConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppli_ds" {
  device_name = "testacc_dataAppli_ds"
  host        = "192.168.100.13"
}

resource "wallix-bastion_device_service" "testacc_dataAppli_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppli_ds.id
  service_name      = "testacc_dataAppli_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppli_ds" {
  cluster_name = "testacc_dataAppli_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppli_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppli_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppli_ds" {
  application_name  = "testacc_dataAppli_ds"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppli_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppli_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppli_ds.cluster_name
  tags {
    key   = "testkey"
    value = "testvalue"
  }
  tags {
    key   = "testkey2"
    value = "testvalue2"
  }
}

data "wallix-bastion_application" "testacc_dataAppli" {
  application_name = wallix-bastion_application.testacc_dataAppli_ds.application_name
}
`
}
