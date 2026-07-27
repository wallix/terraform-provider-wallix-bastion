package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// certificate_authorities doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceCertificateAuthority_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			ExternalProviders: map[string]resource.ExternalProvider{
				"tls": {
					Source: "hashicorp/tls",
				},
			},
			Steps: []resource.TestStep{
				{
					Config: testAccResourceCertificateAuthorityCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_certificate_authority.testacc_CertificateAuthority",
							"id"),
						resource.TestCheckResourceAttr(
							"wallix-bastion_certificate_authority.testacc_CertificateAuthority",
							"ca_type", "X509"),
					),
				},
				{
					Config: testAccResourceCertificateAuthorityUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"wallix-bastion_certificate_authority.testacc_CertificateAuthority",
							"description", "testacc updated description"),
					),
				},
				{
					ResourceName:      "wallix-bastion_certificate_authority.testacc_CertificateAuthority",
					ImportState:       true,
					ImportStateVerify: false,
					ImportStateId:     "testacc_CertificateAuthority",
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

func testAccResourceCertificateAuthorityCreate() string {
	return `
resource "tls_private_key" "testacc_CertificateAuthority" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "testacc_CertificateAuthority" {
  private_key_pem = tls_private_key.testacc_CertificateAuthority.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA"
    organization = "WALLIX Test"
  }

  validity_period_hours = 8760

  is_ca_certificate = true

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
    "crl_signing",
  ]
}

resource "wallix-bastion_certificate_authority" "testacc_CertificateAuthority" {
  certificate_authority_name = "testacc_CertificateAuthority"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_CertificateAuthority.cert_pem
}
`
}

func testAccResourceCertificateAuthorityUpdate() string {
	return `
resource "tls_private_key" "testacc_CertificateAuthority" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "testacc_CertificateAuthority" {
  private_key_pem = tls_private_key.testacc_CertificateAuthority.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA"
    organization = "WALLIX Test"
  }

  validity_period_hours = 8760

  is_ca_certificate = true

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
    "crl_signing",
  ]
}

resource "wallix-bastion_certificate_authority" "testacc_CertificateAuthority" {
  certificate_authority_name = "testacc_CertificateAuthority"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_CertificateAuthority.cert_pem
  description                = "testacc updated description"
}
`
}
