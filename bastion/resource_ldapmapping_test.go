package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceLDAPMapping_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLDAPMappingCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_ldapmapping.testacc_LDAPMapping",
						"id"),
				),
			},
			{
				ResourceName:  "wallix-bastion_ldapmapping.testacc_LDAPMapping",
				ImportState:   true,
				ImportStateId: "testacc_LDAPMapping/testacc_LDAPMapping/CN=testacc,OU=FR,DC=test,DC=com",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceLDAPMappingCreate() string {
	return `
resource wallix-bastion_ldapdomain testacc_LDAPMapping {
  domain_name          = "testacc_LDAPMapping"
  ldap_domain_name     = "test.com"
  external_ldaps       = [wallix-bastion_externalauth_ldap.testacc_LDAPMapping.authentication_name]
  default_language     = "fr"
  default_email_domain = "test.com"
}
resource wallix-bastion_externalauth_ldap testacc_LDAPMapping {
  authentication_name = "testacc_LDAPMapping"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource wallix-bastion_usergroup testacc_LDAPMapping {
  group_name = "testacc_LDAPMapping"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource wallix-bastion_ldapmapping testacc_LDAPMapping {
  domain     = wallix-bastion_ldapdomain.testacc_LDAPMapping.domain_name
  user_group = wallix-bastion_usergroup.testacc_LDAPMapping.group_name
  ldap_group = "CN=testacc,OU=FR,DC=test,DC=com"
}
`
}
