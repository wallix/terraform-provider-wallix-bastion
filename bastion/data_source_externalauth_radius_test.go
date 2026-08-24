package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExternalAuthRadius_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceExternalAuthRadiusConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceExternalAuthRadiusConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"authentication_name", "testacc_dataExternalAuthRadius"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"host", "192.168.100.20"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"port", "1813"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"timeout", "10"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"description", "testacc ExternalAuthRadius"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius",
						"use_primary_auth_domain", "true"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceExternalAuthRadiusConfigCreate() string {
	return `
resource "wallix-bastion_externalauth_radius" "testacc_dataExternalAuthRadius" {
  authentication_name     = "testacc_dataExternalAuthRadius"
  host                    = "192.168.100.20"
  port                    = 1813
  secret                  = "aSecret"
  timeout                 = 10
  description             = "testacc ExternalAuthRadius"
  use_primary_auth_domain = true
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceExternalAuthRadiusConfigData() string {
	return `
resource "wallix-bastion_externalauth_radius" "testacc_dataExternalAuthRadius" {
  authentication_name     = "testacc_dataExternalAuthRadius"
  host                    = "192.168.100.20"
  port                    = 1813
  secret                  = "aSecret"
  timeout                 = 10
  description             = "testacc ExternalAuthRadius"
  use_primary_auth_domain = true
}

data "wallix-bastion_externalauth_radius" "testacc_dataExternalAuthRadius" {
  authentication_name = wallix-bastion_externalauth_radius.testacc_dataExternalAuthRadius.authentication_name
}
`
}
