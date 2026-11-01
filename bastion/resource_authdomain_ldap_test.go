package bastion_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func TestAccResourceAuthDomainLdap_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
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

// certificate_authorities doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceAuthDomainLdap_certificateAuthority(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" && semver.Compare(v, bastion.VersionWallixAPI312) < 0 {
		t.Skip("certificate_authority requires API v3.12 or later")
	}
	const res = "wallix-bastion_authdomain_ldap.testacc_AuthDomainLdapCA"
	const data = "data.wallix-bastion_authdomain_ldap.testacc_AuthDomainLdapCA"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			tvTLSProviderName: {
				Source: tvTLSProviderSource,
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAuthDomainLdapCAConfig(`
  certificate_authority = wallix-bastion_certificate_authority.testacc_AuthDomainLdapCA[0].certificate_authority_name
  enable_ca             = true`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "certificate_authority", "testacc_AuthDomainLdapCA0"),
					resource.TestCheckResourceAttr(res, "enable_ca", "true"),
				),
			},
			{
				Config: testAccResourceAuthDomainLdapCAConfig(`
  certificate_authority = wallix-bastion_certificate_authority.testacc_AuthDomainLdapCA[1].certificate_authority_name
`) + `
data "wallix-bastion_authdomain_ldap" "testacc_AuthDomainLdapCA" {
  domain_name = wallix-bastion_authdomain_ldap.testacc_AuthDomainLdapCA.domain_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "certificate_authority", "testacc_AuthDomainLdapCA1"),
					resource.TestCheckResourceAttr(res, "enable_ca", "false"),
					resource.TestCheckResourceAttr(data, "certificate_authority", "testacc_AuthDomainLdapCA1"),
					resource.TestCheckResourceAttr(data, "enable_ca", "false"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     "testacc.AuthDomainLdapCA",
				ImportStateVerify: true,
			},
			{
				Config:      testAccResourceAuthDomainLdapCAConfig(`  enable_ca = true`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`enable_ca = true requires certificate_authority`),
			},
			{
				// The CAs stay: the Bastion refuses to delete a CA that a domain references, and
				// Terraform doesn't order the domain update before the CA destroy.
				Config: testAccResourceAuthDomainLdapCAConfig(``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "certificate_authority", ""),
					resource.TestCheckResourceAttr(res, "enable_ca", "false"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// API v3.8 refuses certificate_authority and enable_ca: refused at plan.
func TestAccResourceAuthDomainLdap_certificateAuthorityV38(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		t.Skip("only for API versions before v3.12")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainLdapCA" {
  domain_name           = "testacc.AuthDomainLdapCA"
  auth_domain_name      = "testca.com"
  external_auths        = ["testacc_AuthDomainLdapCA"]
  default_language      = "en"
  default_email_domain  = "testca.com"
  certificate_authority = "testacc_AuthDomainLdapCA"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`certificate_authority and enable_ca are not available with API version`),
			},
		},
	})
}

// testAccResourceAuthDomainLdapCAConfig declares two X509 certificate authorities and an LDAP
// domain with the given extra arguments.
func testAccResourceAuthDomainLdapCAConfig(domainArgs string) string {
	return `
resource "tls_private_key" "testacc_AuthDomainLdapCA" {
  algorithm = "RSA"
  rsa_bits  = 2048
}
resource "tls_self_signed_cert" "testacc_AuthDomainLdapCA" {
  count                 = 2
  private_key_pem       = tls_private_key.testacc_AuthDomainLdapCA.private_key_pem
  validity_period_hours = 24
  is_ca_certificate     = true
  allowed_uses          = ["cert_signing", "crl_signing"]
  subject {
    common_name = "testacc AuthDomainLdapCA ${count.index}"
  }
}
resource "wallix-bastion_certificate_authority" "testacc_AuthDomainLdapCA" {
  count                      = 2
  certificate_authority_name = "testacc_AuthDomainLdapCA${count.index}"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_AuthDomainLdapCA[count.index].cert_pem
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainLdapCA" {
  authentication_name = "testacc_AuthDomainLdapCA"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.21"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainLdapCA" {
  domain_name          = "testacc.AuthDomainLdapCA"
  auth_domain_name     = "testca.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainLdapCA.authentication_name]
  default_language     = "en"
  default_email_domain = "testca.com"
` + domainArgs + `
}
`
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
  host                = "192.168.100.21"
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
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
`
}
