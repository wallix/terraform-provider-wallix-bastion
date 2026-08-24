package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExternalAuthTacacs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceExternalAuthTacacsConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceExternalAuthTacacsConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs",
						"authentication_name", "testacc_dataExternalAuthTacacs"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs",
						"host", "192.168.100.20"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs",
						"port", "4949"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs",
						"description", "testacc ExternalAuthTacacs"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs",
						"use_primary_auth_domain", "true"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceExternalAuthTacacsConfigCreate() string {
	return `
resource "wallix-bastion_externalauth_tacacs" "testacc_dataExternalAuthTacacs" {
  authentication_name     = "testacc_dataExternalAuthTacacs"
  host                    = "192.168.100.20"
  port                    = 4949
  secret                  = "aSecret"
  description             = "testacc ExternalAuthTacacs"
  use_primary_auth_domain = true
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceExternalAuthTacacsConfigData() string {
	return `
resource "wallix-bastion_externalauth_tacacs" "testacc_dataExternalAuthTacacs" {
  authentication_name     = "testacc_dataExternalAuthTacacs"
  host                    = "192.168.100.20"
  port                    = 4949
  secret                  = "aSecret"
  description             = "testacc ExternalAuthTacacs"
  use_primary_auth_domain = true
}

data "wallix-bastion_externalauth_tacacs" "testacc_dataExternalAuthTacacs" {
  authentication_name = wallix-bastion_externalauth_tacacs.testacc_dataExternalAuthTacacs.authentication_name
}
`
}
