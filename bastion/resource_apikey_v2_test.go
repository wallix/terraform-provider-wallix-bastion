package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// apikeys-v2 doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceAPIKeyV2_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceAPIKeyV2Create(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_apikey_v2.testacc_APIKeyV2",
							"id"),
						resource.TestCheckResourceAttr(
							"wallix-bastion_apikey_v2.testacc_APIKeyV2",
							"profile", "user"),
					),
				},
				{
					Config: testAccResourceAPIKeyV2Update(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"wallix-bastion_apikey_v2.testacc_APIKeyV2",
							"description", "testacc updated description"),
					),
				},
				{
					ResourceName:  "wallix-bastion_apikey_v2.testacc_APIKeyV2",
					ImportState:   true,
					ImportStateId: "testacc_APIKeyV2",
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

func testAccResourceAPIKeyV2Create() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_APIKeyV2" {
  apikey_name = "testacc_APIKeyV2"
  profile     = "user"
}
`
}

func testAccResourceAPIKeyV2Update() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_APIKeyV2" {
  apikey_name = "testacc_APIKeyV2"
  profile     = "user"
  description = "testacc updated description"
}
`
}
