package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUserCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_user.testacc_User",
						"id"),
				),
			},
			{
				Config: testAccResourceUserUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_user.testacc_User",
				ImportState:   true,
				ImportStateId: "testacc_User",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceUserCreate() string {
	return `
resource wallix-bastion_usergroup testacc_User {
  group_name = "testacc_User"
  timeframes = ["allthetime"]
}
resource "random_password" "testacc_User" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}

resource wallix-bastion_user testacc_User {
  user_name  = "testacc_User"
  email      = "testacc-user@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  groups = [
    wallix-bastion_usergroup.testacc_User.group_name,
  ]
  force_change_pwd   = true
  preferred_language = "fr"
  password           = random_password.testacc_User.result
}
`
}

func testAccResourceUserUpdate() string {
	return `
resource wallix-bastion_usergroup testacc_User {
  group_name = "testacc_User"
  timeframes = ["allthetime"]
}

resource "tls_private_key" "testacc_User" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource wallix-bastion_user testacc_User {
  user_name  = "testacc_User"
  email      = "testacc-user@none.none"
  profile    = "user"
  user_auths = ["local_password", "local_sshkey"]
  groups = [
    wallix-bastion_usergroup.testacc_User.group_name,
  ]
  display_name    = "testacc User"
  expiration_date = "2032-01-03 00:01"
  ip_source       = "127.0.0.1"
  is_disabled     = true
  ssh_public_key  = tls_private_key.testacc_User.public_key_openssh
}
`
}
