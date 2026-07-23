package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// certificate_authorities doesn't exist before API v3.12; skip on older/default versions.
func TestAccDataSourceCertificateAuthority_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			ExternalProviders: map[string]resource.ExternalProvider{
				"tls": {
					Source: "hashicorp/tls",
				},
			},
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceCertificateAuthorityConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceCertificateAuthorityConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_certificate_authority.testacc_dataCertificateAuthority",
							"certificate_authority_name", "testacc_dataCertificateAuthority_ds"),
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_certificate_authority.testacc_dataCertificateAuthority",
							"ca_type", "X509"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_certificate_authority.testacc_dataCertificateAuthority",
							"ca_certificate"),
					),
				},
			},
		})
	}
}

// Resource creation configuration.
func testAccDataSourceCertificateAuthorityConfigCreate() string {
	return `
resource "tls_private_key" "testacc_dataCertificateAuthority_ds" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "testacc_dataCertificateAuthority_ds" {
  private_key_pem = tls_private_key.testacc_dataCertificateAuthority_ds.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA DS"
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

resource "wallix-bastion_certificate_authority" "testacc_dataCertificateAuthority_ds" {
  certificate_authority_name = "testacc_dataCertificateAuthority_ds"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_dataCertificateAuthority_ds.cert_pem
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceCertificateAuthorityConfigData() string {
	return `
resource "tls_private_key" "testacc_dataCertificateAuthority_ds" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "testacc_dataCertificateAuthority_ds" {
  private_key_pem = tls_private_key.testacc_dataCertificateAuthority_ds.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA DS"
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

resource "wallix-bastion_certificate_authority" "testacc_dataCertificateAuthority_ds" {
  certificate_authority_name = "testacc_dataCertificateAuthority_ds"
  ca_type                    = "X509"
  ca_certificate             = tls_self_signed_cert.testacc_dataCertificateAuthority_ds.cert_pem
}

data "wallix-bastion_certificate_authority" "testacc_dataCertificateAuthority" {
  certificate_authority_name = "testacc_dataCertificateAuthority_ds"

  depends_on = [wallix-bastion_certificate_authority.testacc_dataCertificateAuthority_ds]
}
`
}
