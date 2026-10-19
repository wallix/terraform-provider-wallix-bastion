package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceProfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceProfileCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_profile.testacc_Profile",
						"id"),
				),
			},
			{
				Config: testAccResourceProfileUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_profile.testacc_Profile",
				ImportState:   true,
				ImportStateId: "testacc_Profile",
				// The import used to set profile_name to the API id, which forced a replacement.
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if got := states[0].Attributes["profile_name"]; got != "testacc_Profile" {
						return fmt.Errorf("profile_name = %q after import, want testacc_Profile", got)
					}

					return nil
				},
			},
			{
				Config: testAccResourceProfileUpdate2(),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceProfileCreate() string {
	return `
resource "wallix-bastion_profile" "testacc_Profile" {
  profile_name = "testacc_Profile"
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

func testAccResourceProfileUpdate() string {
	return `
resource "wallix-bastion_profile" "testacc_Profile" {
  profile_name  = "testacc_Profile"
  description   = "testacc Profile"
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
  user_groups_limitation {
    user_groups = [wallix-bastion_usergroup.testacc_Profile.group_name]
  }
  target_groups_limitation {
    default_target_group = wallix-bastion_targetgroup.testacc_Profile.group_name
    target_groups        = [wallix-bastion_targetgroup.testacc_Profile.group_name]
  }
}
resource "wallix-bastion_usergroup" "testacc_Profile" {
  group_name = "testacc_Profile"
  timeframes = ["allthetime"]
}
resource "wallix-bastion_targetgroup" "testacc_Profile" {
  group_name = "testacc_Profile"
}
`
}

func testAccResourceProfileUpdate2() string {
	return `
resource "wallix-bastion_profile" "testacc_Profile" {
  profile_name  = "testacc_Profile"
  description   = "testacc Profile"
  ip_limitation = "127.0.0.1"
  target_access = true
  dashboards = [
    "opsadmin",
    "audit",
  ]
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
  user_groups_limitation {
    user_groups = [wallix-bastion_usergroup.testacc_Profile.group_name]
  }
  target_groups_limitation {
    default_target_group = wallix-bastion_targetgroup.testacc_Profile.group_name
    target_groups        = [wallix-bastion_targetgroup.testacc_Profile.group_name]
  }
}
resource "wallix-bastion_usergroup" "testacc_Profile" {
  group_name = "testacc_Profile"
  timeframes = ["allthetime"]
}
resource "wallix-bastion_targetgroup" "testacc_Profile" {
  group_name = "testacc_Profile"
}
`
}
