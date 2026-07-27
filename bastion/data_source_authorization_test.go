package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// The update step (in the sibling resource test) sets authorize_session_sharing, which
// requires API v3.12+; gate this datasource test the same way for consistency.
func TestAccDataSourceAuthorization_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			Providers:                 testAccProviders,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceAuthorizationConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceAuthorizationConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.wallix-bastion_authorization.testacc_dataAuthorization",
							"authorization_name", "testacc_dataAuthorization_ds"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authorization.testacc_dataAuthorization",
							"user_group", "testacc_dataAuthorization_ds_ug"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authorization.testacc_dataAuthorization",
							"target_group", "testacc_dataAuthorization_ds_tg"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authorization.testacc_dataAuthorization",
							"authorize_sessions", "true"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authorization.testacc_dataAuthorization",
							"subprotocols.#", "2"),
					),
				},
			},
		})
	}
}

// Resource creation configuration.
func testAccDataSourceAuthorizationConfigCreate() string {
	return `
resource "wallix-bastion_authorization" "testacc_dataAuthorization_ds" {
  authorization_name = "testacc_dataAuthorization_ds"
  user_group         = wallix-bastion_usergroup.testacc_dataAuthorization_ds.group_name
  target_group       = wallix-bastion_targetgroup.testacc_dataAuthorization_ds.group_name
  authorize_sessions = true
  subprotocols = [
    "RDP",
    "SSH_SHELL_SESSION",
  ]
}

resource "wallix-bastion_usergroup" "testacc_dataAuthorization_ds" {
  group_name = "testacc_dataAuthorization_ds_ug"
  timeframes = ["allthetime"]
}

resource "wallix-bastion_targetgroup" "testacc_dataAuthorization_ds" {
  group_name = "testacc_dataAuthorization_ds_tg"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthorizationConfigData() string {
	return `
resource "wallix-bastion_authorization" "testacc_dataAuthorization_ds" {
  authorization_name = "testacc_dataAuthorization_ds"
  user_group         = wallix-bastion_usergroup.testacc_dataAuthorization_ds.group_name
  target_group       = wallix-bastion_targetgroup.testacc_dataAuthorization_ds.group_name
  authorize_sessions = true
  subprotocols = [
    "RDP",
    "SSH_SHELL_SESSION",
  ]
}

resource "wallix-bastion_usergroup" "testacc_dataAuthorization_ds" {
  group_name = "testacc_dataAuthorization_ds_ug"
  timeframes = ["allthetime"]
}

resource "wallix-bastion_targetgroup" "testacc_dataAuthorization_ds" {
  group_name = "testacc_dataAuthorization_ds_tg"
}

data "wallix-bastion_authorization" "testacc_dataAuthorization" {
  authorization_name = wallix-bastion_authorization.testacc_dataAuthorization_ds.authorization_name
}
`
}
