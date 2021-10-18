package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthTacacs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
resource wallix-bastion_externalauth_tacacs testacc_ExternalAuthTacacs {
  authentication_name = "testacc_ExternalAuthTacacs"
  host                = "server1"
  port                = 49
  secret              = "aSecret"
}
`
}

func testAccResourceExternalAuthTacacsUpdate() string {
	return `
resource wallix-bastion_externalauth_tacacs testacc_ExternalAuthTacacs {
  authentication_name     = "testacc_ExternalAuthTacacs"
  host                    = "server1"
  port                    = 1813
  secret                  = "aSecret"
  description             = "testacc ExternalAuthTacacs"
  use_primary_auth_domain = true
}
`
}
