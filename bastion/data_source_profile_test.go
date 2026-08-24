package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceProfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceProfileConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceProfileConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"profile_name", "testacc_dataProfile"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"description", "testacc data Profile"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"ip_limitation", "127.0.0.1"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"target_access", "true"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"gui_features.0.wab_audit", "view"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"gui_features.0.users", "view"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"gui_transmission.0.users", "view"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"target_groups_limitation.#", "0"),
					resource.TestCheckResourceAttr("data.wallix-bastion_profile.testacc_dataProfile",
						"user_groups_limitation.#", "0"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceProfileConfigCreate() string {
	return `
resource "wallix-bastion_profile" "testacc_dataProfile" {
  profile_name  = "testacc_dataProfile"
  description   = "testacc data Profile"
  ip_limitation = "127.0.0.1"
  target_access = true
  gui_features {
    wab_audit      = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    system_audit   = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
  gui_transmission {
    system_audit   = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceProfileConfigData() string {
	return `
resource "wallix-bastion_profile" "testacc_dataProfile" {
  profile_name  = "testacc_dataProfile"
  description   = "testacc data Profile"
  ip_limitation = "127.0.0.1"
  target_access = true
  gui_features {
    wab_audit      = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    system_audit   = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
  gui_transmission {
    system_audit   = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
}

data "wallix-bastion_profile" "testacc_dataProfile" {
  profile_name = wallix-bastion_profile.testacc_dataProfile.profile_name
}
`
}
