package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthRadius_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthRadiusCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_radius.testacc_ExternalAuthRadius",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthRadiusUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_radius.testacc_ExternalAuthRadius",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthRadius",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthRadiusCreate() string {
	return `
resource wallix-bastion_externalauth_radius testacc_ExternalAuthRadius {
  authentication_name = "testacc_ExternalAuthRadius"
  host                = "server1"
  port                = 1813
  secret              = "aSecret"
  timeout             = 10
}
`
}

func testAccResourceExternalAuthRadiusUpdate() string {
	return `
resource wallix-bastion_externalauth_radius testacc_ExternalAuthRadius {
  authentication_name     = "testacc_ExternalAuthRadius"
  host                    = "server1"
  port                    = 1813
  secret                  = "aSecret"
  timeout                 = 10
  description             = "testacc ExternalAuthRadius"
  use_primary_auth_domain = true
}
`
}
