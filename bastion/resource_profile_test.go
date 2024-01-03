package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func TestAccResourceProfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		Steps:                     testAccResourceProfileSteps(),
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceProfileSteps() []resource.TestStep {
	steps := []resource.TestStep{
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
		},
	}
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" &&
		v != bastion.VersionWallixAPI33 {
		steps = append(steps, resource.TestStep{
			Config: testAccResourceProfileUpdate2(),
		})
	}

	return steps
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
