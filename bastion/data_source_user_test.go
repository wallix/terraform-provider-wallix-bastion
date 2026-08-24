package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceUserConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				// The password field is never returned by the API, so it is not checked here.
				Config: testAccDataSourceUserConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"user_name", "testacc_dataUser"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"email", "testacc-datauser@none.none"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"profile", "user"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"user_auths.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_user.testacc_dataUser",
						"user_auths.*", "local_password"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"display_name", "testacc data User"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"preferred_language", "fr"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"is_disabled", "false"),
					resource.TestCheckResourceAttr("data.wallix-bastion_user.testacc_dataUser",
						"groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_user.testacc_dataUser",
						"groups.*", "testacc_dataUser"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceUserConfigCreate() string {
	return `
resource "wallix-bastion_usergroup" "testacc_dataUser" {
  group_name = "testacc_dataUser"
  timeframes = ["allthetime"]
}
resource "random_password" "testacc_dataUser" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}

resource "wallix-bastion_user" "testacc_dataUser" {
  user_name  = "testacc_dataUser"
  email      = "testacc-datauser@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  groups = [
    wallix-bastion_usergroup.testacc_dataUser.group_name,
  ]
  display_name       = "testacc data User"
  preferred_language = "fr"
  password           = random_password.testacc_dataUser.result
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceUserConfigData() string {
	return `
resource "wallix-bastion_usergroup" "testacc_dataUser" {
  group_name = "testacc_dataUser"
  timeframes = ["allthetime"]
}
resource "random_password" "testacc_dataUser" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}

resource "wallix-bastion_user" "testacc_dataUser" {
  user_name  = "testacc_dataUser"
  email      = "testacc-datauser@none.none"
  profile    = "user"
  user_auths = ["local_password"]
  groups = [
    wallix-bastion_usergroup.testacc_dataUser.group_name,
  ]
  display_name       = "testacc data User"
  preferred_language = "fr"
  password           = random_password.testacc_dataUser.result
}

data "wallix-bastion_user" "testacc_dataUser" {
  user_name = wallix-bastion_user.testacc_dataUser.user_name
}
`
}
