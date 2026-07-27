package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// config_smtp is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "smtpConfig"). The data
// source therefore has no Required lookup field: it always reads the single
// existing SMTP configuration, so the "data" block below takes no arguments.
func TestAccDataSourceConfigSMTP_basic(t *testing.T) {
	dataSourceName := "data.wallix-bastion_config_smtp.current"

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConfigSMTPConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConfigSMTPConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "smtp"),
					resource.TestCheckResourceAttr(dataSourceName, "server", "testacc-smtp-ds.example.com"),
					resource.TestCheckResourceAttr(dataSourceName, "port", "25"),
				),
			},
		},
	})
}

func testAccDataSourceConfigSMTPConfigCreate() string {
	return `
resource "wallix-bastion_config_smtp" "test_ds" {
  protocol               = "smtp"
  authentication_method  = "off"
  server                 = "testacc-smtp-ds.example.com"
  port                   = 25
  postmaster_email       = "postmaster@testacc.example.com"
  sender_name            = "WALLIX Bastion"
  sender_email           = "bastion@testacc.example.com"
}
`
}

func testAccDataSourceConfigSMTPConfigData() string {
	return testAccDataSourceConfigSMTPConfigCreate() + `
data "wallix-bastion_config_smtp" "current" {}
`
}
