package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// apikeys-v2 doesn't exist before API v3.12; skip on older/default versions.
func TestAccDataSourceAPIKeyV2_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			Providers:                 testAccProviders,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceAPIKeyV2ConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceAPIKeyV2ConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.wallix-bastion_apikey_v2.testacc_dataAPIKeyV2",
							"apikey_name", "testacc_dataAPIKeyV2_ds"),
						resource.TestCheckResourceAttr("data.wallix-bastion_apikey_v2.testacc_dataAPIKeyV2",
							"profile", "user"),
						resource.TestCheckResourceAttrSet("data.wallix-bastion_apikey_v2.testacc_dataAPIKeyV2",
							"apikey"),
					),
				},
			},
		})
	}
}

// Resource creation configuration.
func testAccDataSourceAPIKeyV2ConfigCreate() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_dataAPIKeyV2_ds" {
  apikey_name = "testacc_dataAPIKeyV2_ds"
  profile     = "user"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAPIKeyV2ConfigData() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_dataAPIKeyV2_ds" {
  apikey_name = "testacc_dataAPIKeyV2_ds"
  profile     = "user"
}

data "wallix-bastion_apikey_v2" "testacc_dataAPIKeyV2" {
  apikey_name = wallix-bastion_apikey_v2.testacc_dataAPIKeyV2_ds.apikey_name
}
`
}
