package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceClusterConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceClusterConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_cluster.testacc_dataCluster",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_cluster.testacc_dataCluster",
						"cluster_name", "testacc_Cluster_ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_cluster.testacc_dataCluster",
						"interactive_logins.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_cluster.testacc_dataCluster",
						"interactive_logins.*", "testacc_Cluster_ds:testacc_Cluster_ds"),
				),
			},
		},
	})
}

// nolint: lll, nolintlint
// Resource creation configuration.
func testAccDataSourceClusterConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_Cluster_ds" {
  device_name = "testacc_Cluster_ds"
  host        = "192.168.100.12"
}

resource "wallix-bastion_device_service" "testacc_Cluster_ds" {
  device_id         = wallix-bastion_device.testacc_Cluster_ds.id
  service_name      = "testacc_Cluster_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN"]
}

resource "wallix-bastion_cluster" "testacc_Cluster_ds" {
  cluster_name = "testacc_Cluster_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_Cluster_ds.device_name}:${wallix-bastion_device_service.testacc_Cluster_ds.service_name}",
  ]
}
`
}

// nolint: lll, nolintlint
// Datasource configuration to retrieve the created resource.
func testAccDataSourceClusterConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_Cluster_ds" {
  device_name = "testacc_Cluster_ds"
  host        = "192.168.100.12"
}

resource "wallix-bastion_device_service" "testacc_Cluster_ds" {
  device_id         = wallix-bastion_device.testacc_Cluster_ds.id
  service_name      = "testacc_Cluster_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN"]
}

resource "wallix-bastion_cluster" "testacc_Cluster_ds" {
  cluster_name = "testacc_Cluster_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_Cluster_ds.device_name}:${wallix-bastion_device_service.testacc_Cluster_ds.service_name}",
  ]
}

data "wallix-bastion_cluster" "testacc_dataCluster" {
  cluster_name = wallix-bastion_cluster.testacc_Cluster_ds.cluster_name
}
`
}
