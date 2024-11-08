package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuthDomainAD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAuthDomainADCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_authdomain_ad.testacc_AuthDomainAD",
						"id"),
				),
			},
			{
				Config: testAccResourceAuthDomainADUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_authdomain_ad.testacc_AuthDomainAD",
				ImportState:   true,
				ImportStateId: "testacc.AuthDomainAD-u",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceAuthDomainADCreate() string {
	return `
resource "wallix-bastion_authdomain_ad" "testacc_AuthDomainAD" {
  domain_name          = "testacc.AuthDomainAD"
  auth_domain_name     = "test2.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainAD.authentication_name]
  default_language     = "fr"
  default_email_domain = "test2.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainAD" {
  authentication_name = "testacc_AuthDomainAD"
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

func testAccResourceAuthDomainADUpdate() string {
	return `
resource "wallix-bastion_authdomain_ad" "testacc_AuthDomainAD" {
  domain_name            = "testacc.AuthDomainAD-u"
  auth_domain_name       = "test2u.com"
  external_auths         = [wallix-bastion_externalauth_ldap.testacc_AuthDomainAD.authentication_name]
  default_language       = "en"
  default_email_domain   = "test2u.com"
  description            = "testacc AuthDomainAD"
  display_name_attribute = "displayName"
  language_attribute     = "preferredLanguage"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainAD" {
  authentication_name = "testacc_ADDomain"
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
