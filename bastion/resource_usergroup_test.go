package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUserGroup_basic(t *testing.T) {
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
				Config: testAccResourceUserGroupCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_usergroup.testacc_Usergroup",
						"id"),
				),
			},
			{
				Config: testAccResourceUserGroupUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_usergroup.testacc_Usergroup",
				ImportState:   true,
				ImportStateId: "testacc_Usergroup",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceUserGroupCreate() string {
	return `
resource "random_password" "testacc_Usergroup" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource wallix-bastion_user testacc_Usergroup {
  user_name  = "testacc_Usergroup"
  email      = "testacc-usergroup@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  password   = random_password.testacc_Usergroup.result
}
resource wallix-bastion_usergroup testacc_Usergroup {
  group_name = "testacc_Usergroup"
  timeframes = ["allthetime"]
  users = [
    wallix-bastion_user.testacc_Usergroup.user_name
  ]
}
`
}

func testAccResourceUserGroupUpdate() string {
	return `
resource "random_password" "testacc_Usergroup" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource wallix-bastion_user testacc_Usergroup {
  user_name  = "testacc_Usergroup"
  email      = "testacc-usergroup@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  password   = random_password.testacc_Usergroup.result
}
resource wallix-bastion_usergroup testacc_Usergroup {
  group_name  = "testacc_Usergroup"
  timeframes  = ["allthetime"]
  description = "testacc User Group"
  profile     = "user"
  restrictions {
    action      = "notify"
    rules       = "sudo"
    subprotocol = "SSH_SHELL_SESSION"
  }
  users = [
    wallix-bastion_user.testacc_Usergroup.user_name
  ]
}
`
}
