package bastion_test

import (
	"os"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceLDAPDomain_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" ||
		v == bastion.VersionWallixAPI33 ||
		v == bastion.VersionWallixAPI36 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceLDAPDomainCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_ldapdomain.testacc_LDAPDomain",
							"id"),
					),
				},
				{
					Config: testAccResourceLDAPDomainUpdate(),
				},
				{
					ResourceName:  "wallix-bastion_ldapdomain.testacc_LDAPDomain",
					ImportState:   true,
					ImportStateId: "testacc_LDAPDomain",
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

func testAccResourceLDAPDomainCreate() string {
	return `
resource "wallix-bastion_ldapdomain" "testacc_LDAPDomain" {
  domain_name          = "testacc_LDAPDomain"
  ldap_domain_name     = "test.com"
  external_ldaps       = [wallix-bastion_externalauth_ldap.testacc_LDAPDomain.authentication_name]
  default_language     = "fr"
  default_email_domain = "test.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_LDAPDomain" {
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

func testAccResourceLDAPDomainUpdate() string {
	return `
resource "wallix-bastion_ldapdomain" "testacc_LDAPDomain" {
  domain_name            = "testacc_LDAPDomain"
  ldap_domain_name       = "test.com"
  external_ldaps         = [wallix-bastion_externalauth_ldap.testacc_LDAPDomain.authentication_name]
  default_language       = "fr"
  default_email_domain   = "test.com"
  description            = "testacc LDAPDomain"
  display_name_attribute = "displayName"
  language_attribute     = "preferredLanguage"
}
resource "wallix-bastion_externalauth_ldap" "testacc_LDAPDomain" {
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
