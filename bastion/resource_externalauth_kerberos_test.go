package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthKerberos_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthKerberosCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_kerberos.testacc_ExternalAuthKerberos",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthKerberosUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_kerberos.testacc_ExternalAuthKerberos",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthKerberos",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthKerberosCreate() string {
	return `
resource wallix-bastion_externalauth_kerberos testacc_ExternalAuthKerberos {
  authentication_name = "testacc_ExternalAuthKerberos"
  host                = "server1"
  ker_dom_controller  = "controller"
  port                = 88
}
`
}

func testAccResourceExternalAuthKerberosUpdate() string {
	return `
resource wallix-bastion_externalauth_kerberos testacc_ExternalAuthKerberos {
  authentication_name     = "testacc_ExternalAuthKerberos"
  host                    = "server1"
  ker_dom_controller      = "controller"
  port                    = 88
  kerberos_password       = true
  description             = "testacc ExternalAuthKerberos"
  login_attribute         = "attribute"
  use_primary_auth_domain = true
}
`
}
