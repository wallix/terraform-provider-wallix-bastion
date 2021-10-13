package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClusterCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_cluster.testacc_Cluster",
						"id"),
				),
			},
			{
				Config: testAccResourceClusterUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_cluster.testacc_Cluster",
				ImportState:   true,
				ImportStateId: "testacc_Cluster",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll
func testAccResourceClusterCreate() string {
	return `
resource wallix-bastion_device testacc_Cluster {
  device_name = "testacc_Cluster"
  host        = "testacc_Cluster"
}

resource wallix-bastion_device_service testacc_Cluster {
  device_id         = wallix-bastion_device.testacc_Cluster.id
  service_name      = "testacc_Cluster"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN"]
}

resource wallix-bastion_cluster testacc_Cluster {
  cluster_name = "testacc_Cluster"
  interactive_logins = [
    "${wallix-bastion_device.testacc_Cluster.device_name}:${wallix-bastion_device_service.testacc_Cluster.service_name}",
  ]
}
`
}

// nolint: lll
func testAccResourceClusterUpdate() string {
	return `
resource wallix-bastion_device testacc_Cluster {
  device_name = "testacc_Cluster"
  host        = "testacc_Cluster"
}

resource wallix-bastion_device_service testacc_Cluster {
  device_id         = wallix-bastion_device.testacc_Cluster.id
  service_name      = "testacc_Cluster"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN"]
}
resource wallix-bastion_device_localdomain testacc_Cluster {
  device_id   = wallix-bastion_device.testacc_Cluster.id
  domain_name = "testacc_Cluster"
}
resource wallix-bastion_device_localdomain_account testacc_Cluster {
  device_id     = wallix-bastion_device.testacc_Cluster.id
  domain_id     = wallix-bastion_device_localdomain.testacc_Cluster.id
  account_name  = "testacc_Cluster_admin"
  account_login = "admin"
  services      = [wallix-bastion_device_service.testacc_Cluster.service_name]
}

resource wallix-bastion_cluster testacc_Cluster {
  cluster_name = "testacc_Cluster"
  description  = "testacc Cluster"
  accounts = [
    "${wallix-bastion_device_localdomain_account.testacc_Cluster.account_name}@${wallix-bastion_device_localdomain.testacc_Cluster.domain_name}@${wallix-bastion_device.testacc_Cluster.device_name}:${wallix-bastion_device_service.testacc_Cluster.service_name}",
  ]
  interactive_logins = [
    "${wallix-bastion_device.testacc_Cluster.device_name}:${wallix-bastion_device_service.testacc_Cluster.service_name}",
  ]
}
`
}
