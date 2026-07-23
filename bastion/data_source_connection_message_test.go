package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConnectionMessage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConnectionMessageConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConnectionMessageConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_connection_message.testacc_dataConnectionMessage",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_connection_message.testacc_dataConnectionMessage",
						"message_name", "login_en"),
					resource.TestCheckResourceAttr("data.wallix-bastion_connection_message.testacc_dataConnectionMessage",
						"message", "TEST ACC DS: message content used by the connection_message datasource test.\n"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceConnectionMessageConfigCreate() string {
	return `
resource "wallix-bastion_connection_message" "testacc_ConnectionMessage_ds" {
  message_name = "login_en"
  message      = "TEST ACC DS: message content used by the connection_message datasource test.\n"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceConnectionMessageConfigData() string {
	return `
resource "wallix-bastion_connection_message" "testacc_ConnectionMessage_ds" {
  message_name = "login_en"
  message      = "TEST ACC DS: message content used by the connection_message datasource test.\n"
}

data "wallix-bastion_connection_message" "testacc_dataConnectionMessage" {
  message_name = wallix-bastion_connection_message.testacc_ConnectionMessage_ds.message_name
}
`
}
