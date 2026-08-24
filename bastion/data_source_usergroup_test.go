package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUserGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			tvRandomProviderName: {
				Source: tvRandomProviderSource,
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceUserGroupConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceUserGroupConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"group_name", "testacc_dataUsergroup"),
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"description", "testacc data User Group"),
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"profile", "user"),
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"timeframes.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"timeframes.*", "allthetime"),
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"restrictions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"restrictions.*", map[string]string{
							"action":      "notify",
							"rules":       "sudo",
							"subprotocol": "SSH_SHELL_SESSION",
						}),
					resource.TestCheckResourceAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"users.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_usergroup.testacc_dataUsergroup",
						"users.*", "testacc_dataUsergroup"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceUserGroupConfigCreate() string {
	return `
resource "random_password" "testacc_dataUsergroup" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "wallix-bastion_user" "testacc_dataUsergroup" {
  user_name  = "testacc_dataUsergroup"
  email      = "testacc-datausergroup@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  password   = random_password.testacc_dataUsergroup.result
}
resource "wallix-bastion_usergroup" "testacc_dataUsergroup" {
  group_name  = "testacc_dataUsergroup"
  timeframes  = ["allthetime"]
  description = "testacc data User Group"
  profile     = "user"
  restrictions {
    action      = "notify"
    rules       = "sudo"
    subprotocol = "SSH_SHELL_SESSION"
  }
  users = [
    wallix-bastion_user.testacc_dataUsergroup.user_name
  ]
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceUserGroupConfigData() string {
	return `
resource "random_password" "testacc_dataUsergroup" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "wallix-bastion_user" "testacc_dataUsergroup" {
  user_name  = "testacc_dataUsergroup"
  email      = "testacc-datausergroup@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  password   = random_password.testacc_dataUsergroup.result
}
resource "wallix-bastion_usergroup" "testacc_dataUsergroup" {
  group_name  = "testacc_dataUsergroup"
  timeframes  = ["allthetime"]
  description = "testacc data User Group"
  profile     = "user"
  restrictions {
    action      = "notify"
    rules       = "sudo"
    subprotocol = "SSH_SHELL_SESSION"
  }
  users = [
    wallix-bastion_user.testacc_dataUsergroup.user_name
  ]
}

data "wallix-bastion_usergroup" "testacc_dataUsergroup" {
  group_name = wallix-bastion_usergroup.testacc_dataUsergroup.group_name
}
`
}
