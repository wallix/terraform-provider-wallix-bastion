package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceConfigSMTP_basic(t *testing.T) {
	resourceName := "wallix-bastion_config_smtp.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigSMTPCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "smtp"),
					resource.TestCheckResourceAttr(resourceName, "server", "testacc-smtp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "25"),
				),
			},
			{
				Config: testAccResourceConfigSMTPUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "protocol", "starttls"),
					resource.TestCheckResourceAttr(resourceName, "port", "587"),
					resource.TestCheckResourceAttr(resourceName, "sender_name", "testacc updated sender"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "smtpConfig",
			},
		},
		PreventPostDestroyRefresh: true, // Config always exists on the Bastion; nothing to destroy.
	})
}

func testAccResourceConfigSMTPCreate() string {
	return `
resource "wallix-bastion_config_smtp" "test" {
  protocol               = "smtp"
  authentication_method  = "off"
  server                 = "testacc-smtp.example.com"
  port                   = 25
  postmaster_email       = "postmaster@testacc.example.com"
  sender_name            = "WALLIX Bastion"
  sender_email           = "bastion@testacc.example.com"
}
`
}

func testAccResourceConfigSMTPUpdate() string {
	return `
resource "wallix-bastion_config_smtp" "test" {
  protocol               = "starttls"
  authentication_method  = "off"
  server                 = "testacc-smtp.example.com"
  port                   = 587
  postmaster_email       = "postmaster@testacc.example.com"
  sender_name            = "testacc updated sender"
  sender_email           = "bastion@testacc.example.com"
}
`
}
