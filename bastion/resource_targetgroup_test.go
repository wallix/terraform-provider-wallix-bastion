package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTargetgroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTargetgroupCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_targetgroup.testacc_Targetgroup",
						"id"),
				),
			},
			{
				Config: testAccResourceTargetgroupUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_targetgroup.testacc_Targetgroup",
				ImportState:   true,
				ImportStateId: "testacc_Targetgroup",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceTargetgroupCreate() string {
	return `
resource wallix-bastion_targetgroup testacc_Targetgroup {
  group_name = "testacc_Targetgroup"
}
`
}

func testAccResourceTargetgroupUpdate() string {
	return `
resource wallix-bastion_targetgroup testacc_Targetgroup {
  group_name  = "testacc_Targetgroup"
  description = "testacc Targetgroup"
  restrictions {
    action      = "notify"
    rules       = "command"
    subprotocol = "SSH_REMOTE_COMMAND"
  }
  password_retrieval_accounts {
    account     = wallix-bastion_domain_account.testacc_Targetgroup.account_name
    domain      = wallix-bastion_domain.testacc_Targetgroup.domain_name
    domain_type = "global"
  }
  session_accounts {
    account     = wallix-bastion_domain_account.testacc_Targetgroup.account_name
    domain      = wallix-bastion_domain.testacc_Targetgroup.domain_name
    domain_type = "global"
    device      = wallix-bastion_device.testacc_Targetgroup.device_name
    service     = wallix-bastion_device_service.testacc_Targetgroup.service_name
  }
  session_accounts {
    account     = wallix-bastion_device_localdomain_account.testacc_Targetgroup2.account_name
    domain      = wallix-bastion_device_localdomain.testacc_Targetgroup2.domain_name
    domain_type = "local"
    device      = wallix-bastion_device.testacc_Targetgroup2.device_name
    service     = wallix-bastion_device_service.testacc_Targetgroup2.service_name
  }
  session_account_mappings {
    device  = wallix-bastion_device.testacc_Targetgroup2.device_name
    service = wallix-bastion_device_service.testacc_Targetgroup2.service_name
  }
  session_interactive_logins {
    device  = wallix-bastion_device.testacc_Targetgroup.device_name
    service = wallix-bastion_device_service.testacc_Targetgroup.service_name
  }
}
resource wallix-bastion_device testacc_Targetgroup {
  device_name = "testacc_Targetgroup"
  host        = "testacc_Targetgroup.device"
}
resource wallix-bastion_device_service testacc_Targetgroup {
  device_id         = wallix-bastion_device.testacc_Targetgroup.id
  service_name      = "testacc_Targetgroup"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
  global_domains    = [wallix-bastion_domain.testacc_Targetgroup.domain_name]
}
resource wallix-bastion_device testacc_Targetgroup2 {
  device_name = "testacc_Targetgroup2"
  host        = "testacc_Targetgroup2.device"
}
resource wallix-bastion_device_service testacc_Targetgroup2 {
  device_id         = wallix-bastion_device.testacc_Targetgroup2.id
  service_name      = "testacc_Targetgroup2"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
}
resource wallix-bastion_device_localdomain testacc_Targetgroup2 {
  device_id   = wallix-bastion_device.testacc_Targetgroup2.id
  domain_name = "testacc_Targetgroup2"
}
resource wallix-bastion_device_localdomain_account testacc_Targetgroup2 {
  device_id     = wallix-bastion_device.testacc_Targetgroup2.id
  domain_id     = wallix-bastion_device_localdomain.testacc_Targetgroup2.id
  account_name  = "testacc_Targetgroup2_admin"
  account_login = "admin"
  services      = [wallix-bastion_device_service.testacc_Targetgroup2.service_name]
}
resource wallix-bastion_domain testacc_Targetgroup {
  domain_name = "testacc_Targetgroup"
}
resource wallix-bastion_domain_account testacc_Targetgroup {
  domain_id     = wallix-bastion_domain.testacc_Targetgroup.id
  account_name  = "testacc_Targetgroup_Admin"
  account_login = "admin"
}
`
}
