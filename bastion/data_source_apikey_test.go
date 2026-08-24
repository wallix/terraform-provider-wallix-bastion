package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAPIKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAPIKeyConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAPIKeyConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_apikey.testacc_dataAPIKey",
						"apikey_name", "testacc_dataAPIKey_ds"),
					resource.TestCheckResourceAttrSet("data.wallix-bastion_apikey.testacc_dataAPIKey",
						"apikey"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAPIKeyConfigCreate() string {
	return `
resource "wallix-bastion_apikey" "testacc_dataAPIKey_ds" {
  apikey_name = "testacc_dataAPIKey_ds"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAPIKeyConfigData() string {
	return `
resource "wallix-bastion_apikey" "testacc_dataAPIKey_ds" {
  apikey_name = "testacc_dataAPIKey_ds"
}

data "wallix-bastion_apikey" "testacc_dataAPIKey" {
  apikey_name = wallix-bastion_apikey.testacc_dataAPIKey_ds.apikey_name
}
`
}
