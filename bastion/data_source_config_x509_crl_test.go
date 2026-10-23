package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConfigX509CRL_basic(t *testing.T) {
	dataSourceName := "data.wallix-bastion_config_x509_crl.current"
	crl := testAccCRL(t, 1)

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories:  testAccProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConfigX509CRLConfigCreate(crl),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConfigX509CRLConfigData(crl),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "x509CRLConfig"),
					resource.TestCheckResourceAttr(dataSourceName, "crl_check", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "crl_auto_fetch", ""),
				),
			},
		},
	})
}

func testAccDataSourceConfigX509CRLConfigCreate(crl string) string {
	return testAccResourceConfigX509CRLFile(crl)
}

func testAccDataSourceConfigX509CRLConfigData(crl string) string {
	return testAccDataSourceConfigX509CRLConfigCreate(crl) + `
data "wallix-bastion_config_x509_crl" "current" {}
`
}
