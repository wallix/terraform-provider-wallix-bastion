package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// config_wsm is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "wsmConfig") and doesn't exist
// before API v3.12; skip on older versions. Default (unset) is v3.12+. The
// data source has no Required lookup field: it always reads the single
// existing WSM configuration, so the "data" block below takes no arguments.
func TestAccDataSourceConfigWSM_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		dataSourceName := "data.wallix-bastion_config_wsm.current"

		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			Providers:                 testAccProviders,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceConfigWSMConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceConfigWSMConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(dataSourceName, "id"),
						resource.TestCheckResourceAttr(dataSourceName, "hostname", "wsm-ds.testacc.example.com"),
						resource.TestCheckResourceAttrSet(dataSourceName, "jws_public"),
					),
				},
			},
		})
	}
}

func testAccDataSourceConfigWSMConfigCreate() string {
	return `
resource "wallix-bastion_config_wsm" "test_ds" {
  hostname = "wsm-ds.testacc.example.com"
}
`
}

func testAccDataSourceConfigWSMConfigData() string {
	return testAccDataSourceConfigWSMConfigCreate() + `
data "wallix-bastion_config_wsm" "current" {}
`
}
