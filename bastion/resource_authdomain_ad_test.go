package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func TestAccResourceAuthDomainAD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
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

// certificate_authorities doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceAuthDomainAD_certificateAuthority(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" && semver.Compare(v, bastion.VersionWallixAPI312) < 0 {
		t.Skip("certificate_authority requires API v3.12 or later")
	}
	const res = "wallix-bastion_authdomain_ad.testacc_AuthDomainADCA"
	const data = "data.wallix-bastion_authdomain_ad.testacc_AuthDomainADCA"
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
				Config: testAccResourceAuthDomainADCAConfig(`
  certificate_authority = wallix-bastion_certificate_authority.testacc_AuthDomainADCA.certificate_authority_name
  enable_ca             = true`) + `
data "wallix-bastion_authdomain_ad" "testacc_AuthDomainADCA" {
  domain_name      = wallix-bastion_authdomain_ad.testacc_AuthDomainADCA.domain_name
  auth_domain_name = wallix-bastion_authdomain_ad.testacc_AuthDomainADCA.auth_domain_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "certificate_authority", "testacc_AuthDomainADCA"),
					resource.TestCheckResourceAttr(res, "enable_ca", "true"),
					resource.TestCheckResourceAttr(data, "certificate_authority", "testacc_AuthDomainADCA"),
					resource.TestCheckResourceAttr(data, "enable_ca", "true"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     "testacc.AuthDomainADCA",
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceAuthDomainADCAConfig(``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "certificate_authority", ""),
					resource.TestCheckResourceAttr(res, "enable_ca", "false"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceAuthDomainADCAConfig(domainArgs string) string {
	return `
resource "tls_private_key" "testacc_AuthDomainADCA" {
  algorithm = "RSA"
  rsa_bits  = 2048
}
resource "tls_self_signed_cert" "testacc_AuthDomainADCA" {
  private_key_pem       = tls_private_key.testacc_AuthDomainADCA.private_key_pem
  validity_period_hours = 24
  is_ca_certificate     = true
  allowed_uses          = ["cert_signing", "crl_signing"]
  subject {
    common_name = "testacc AuthDomainADCA"
  }
}
resource "wallix-bastion_certificate_authority" "testacc_AuthDomainADCA" {
  certificate_authority_name = "testacc_AuthDomainADCA"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_AuthDomainADCA.cert_pem
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainADCA" {
  authentication_name = "testacc_AuthDomainADCA"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.21"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_authdomain_ad" "testacc_AuthDomainADCA" {
  domain_name          = "testacc.AuthDomainADCA"
  auth_domain_name     = "testca.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainADCA.authentication_name]
  default_language     = "en"
  default_email_domain = "testca.com"
` + domainArgs + `
}
`
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
