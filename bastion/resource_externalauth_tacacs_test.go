package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthTacacs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthTacacsCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_tacacs.testacc_ExternalAuthTacacs",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthTacacsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_externalauth_tacacs.testacc_ExternalAuthTacacs",
						"port", "4949"),
				),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_tacacs.testacc_ExternalAuthTacacs",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthTacacs",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthTacacsCreate() string {
	return `
resource "wallix-bastion_externalauth_tacacs" "testacc_ExternalAuthTacacs" {
  authentication_name = "testacc_ExternalAuthTacacs"
  host                = "192.168.100.20"
  port                = 49
  secret              = "aSecret"
}
`
}

func testAccResourceExternalAuthTacacsUpdate() string {
	return `
resource "wallix-bastion_externalauth_tacacs" "testacc_ExternalAuthTacacs" {
  authentication_name     = "testacc_ExternalAuthTacacs"
  host                    = "192.168.100.20"
  port                    = 4949
  secret                  = "aSecret"
  description             = "testacc ExternalAuthTacacs"
  use_primary_auth_domain = true
}
`
}
