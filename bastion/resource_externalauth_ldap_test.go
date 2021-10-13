package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthLDAP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthLDAPCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAP",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthLDAPUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAP",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthLDAP",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthLDAPCreate() string {
	return `
resource wallix-bastion_externalauth_ldap testacc_ExternalAuthLDAP {
  authentication_name = "testacc_ExternalAuthLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
`
}

func testAccResourceExternalAuthLDAPUpdate() string {
	return `
resource wallix-bastion_externalauth_ldap testacc_ExternalAuthLDAP {
  authentication_name = "testacc_ExternalAuthLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_active_directory = true
  login               = "svc1"
  password            = "aPassword"
  description         = "testacc ExternalAuthLDAP"
  is_protected_user   = true
}
`
}
