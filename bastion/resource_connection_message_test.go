package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceConnectionMessage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConnectionMessageCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_connection_message.PrimaryEn",
						"id"),
				),
			},
			{
				Config: testAccResourceConnectionMessageUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_connection_message.SecondaryEn",
				ImportState:   true,
				ImportStateId: "motd_en",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// nolint: lll, nolintlint
func testAccResourceConnectionMessageCreate() string {
	return `
resource "wallix-bastion_connection_message" "PrimaryEn" {
  message_name = "login_en"
  message      = "TEST ACC : WARNING: Access to this system is restricted to duly authorized users only. Any attempt to access this system without authorization or fraudulently remaining within such system will be prosecuted in accordance with the law. Any authorized user is hereby informed and acknowledges that his/her actions may be recorded, retained and audited.\n"
}
resource "wallix-bastion_connection_message" "SecondaryEn" {
  message_name = "motd_en"
  message      = <<EOT
TEST ACC
You are hereby informed and acknowledge that your actions may be recorded, retained and audited in accordance with your organization security policy.
Please contact your WALLIX Bastion administrator for further information.
EOT
}
`
}

// nolint: lll, nolintlint
func testAccResourceConnectionMessageUpdate() string {
	return `
resource "wallix-bastion_connection_message" "PrimaryEn" {
  message_name = "login_en"
  message      = "WARNING: Access to this system is restricted to duly authorized users only. Any attempt to access this system without authorization or fraudulently remaining within such system will be prosecuted in accordance with the law. Any authorized user is hereby informed and acknowledges that his/her actions may be recorded, retained and audited.\n"
}
resource "wallix-bastion_connection_message" "SecondaryEn" {
  message_name = "motd_en"
  message      = <<EOT
You are hereby informed and acknowledge that your actions may be recorded, retained and audited in accordance with your organization security policy.
Please contact your WALLIX Bastion administrator for further information.
EOT
}
`
}
