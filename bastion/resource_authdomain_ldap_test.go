package bastion_test

import (
	"os"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuthDomainLdap_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" &&
		v != bastion.VersionWallixAPI33 &&
		v != bastion.VersionWallixAPI36 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceAuthDomainLdapCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_authdomain_ldap.testacc_AuthDomainLDAP",
							"id"),
					),
				},
				{
					Config: testAccResourceAuthDomainLdapUpdate(),
				},
				{
					ResourceName:  "wallix-bastion_authdomain_ldap.testacc_AuthDomainLDAP",
					ImportState:   true,
					ImportStateId: "testacc.AuthDomainLDAP-u",
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

func testAccResourceAuthDomainLdapCreate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainLDAP" {
  domain_name          = "testacc.AuthDomainLDAP"
  auth_domain_name     = "test3.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainLDAP.authentication_name]
  default_language     = "fr"
  default_email_domain = "test3.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainLDAP" {
  authentication_name = "testacc_AuthDomainLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "server2"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
`
}

func testAccResourceAuthDomainLdapUpdate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainLDAP" {
  domain_name            = "testacc.AuthDomainLDAP-u"
  auth_domain_name       = "test3u.com"
  external_auths         = [wallix-bastion_externalauth_ldap.testacc_AuthDomainLDAP.authentication_name]
  default_language       = "en"
  default_email_domain   = "test3u.com"
  description            = "testacc AuthDomainLDAP"
  display_name_attribute = "displayName"
  language_attribute     = "preferredLanguage"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainLDAP" {
  authentication_name = "testacc_LDAPDomain"
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
