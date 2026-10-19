package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConfigX509OCSP_basic(t *testing.T) {
	dataSourceName := "data.wallix-bastion_config_x509_ocsp.current"

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConfigX509OCSPConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConfigX509OCSPConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "x509OCSPConfig"),
					resource.TestCheckResourceAttr(dataSourceName, "ocsp_check", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "default_ocsp_responder",
						"http://testacc-ocsp-ds.example.com"),
					resource.TestCheckResourceAttr(dataSourceName, "override_ocsp_responder", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "ocsp_proxy", ""),
				),
			},
		},
	})
}

func testAccDataSourceConfigX509OCSPConfigCreate() string {
	return `
resource "wallix-bastion_config_x509_ocsp" "test_ds" {
  ocsp_check              = false
  default_ocsp_responder  = "http://testacc-ocsp-ds.example.com"
  override_ocsp_responder = true
}
`
}

func testAccDataSourceConfigX509OCSPConfigData() string {
	return testAccDataSourceConfigX509OCSPConfigCreate() + `
data "wallix-bastion_config_x509_ocsp" "current" {}
`
}
