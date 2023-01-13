package bastion_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/claranet/terraform-provider-wallix-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalAuthSaml_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" &&
		v != bastion.VersionWallixAPI33 &&
		v != bastion.VersionWallixAPI36 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			ExternalProviders: map[string]resource.ExternalProvider{
				"tls": {
					Source:            "hashicorp/tls",
					VersionConstraint: "~> 4.0",
				},
			},
			Steps: []resource.TestStep{
				{
					Config: testAccResourceExternalAuthSamlCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"id"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"idp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"saml_request_url"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"saml_request_method"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"sp_metadata"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"sp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"sp_assertion_consumer_service"),
						resource.TestCheckResourceAttrSet(
							"wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
							"sp_single_logout_service"),
					),
				},
				{
					Config: testAccResourceExternalAuthSamlUpdate(),
				},
				{
					ResourceName:  "wallix-bastion_externalauth_saml.testacc_ExternalAuthSaml",
					ImportState:   true,
					ImportStateId: "testacc_ExternalAuthSaml",
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

const (
	//nolint: lll
	idpMetadataSAML = `<?xml version="1.0"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" validUntil="2023-01-13T14:54:34Z" cacheDuration="PT1674053674S" entityID="example.com">
  <md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
	<md:KeyDescriptor use="signing">
	  <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
		<ds:X509Data>
		  <ds:X509Certificate>MIICTDCCAbWgAwIBAgIBADANBgkqhkiG9w0BAQ0FADBDMQswCQYDVQQGEwJmcjEMMAoGA1UECAwDSWRmMRAwDgYDVQQKDAdleGFtcGxlMRQwEgYDVQQDDAtleGFtcGxlLmNvbTAeFw0yMzAxMTExNDU0MjVaFw0yNDAxMTExNDU0MjVaMEMxCzAJBgNVBAYTAmZyMQwwCgYDVQQIDANJZGYxEDAOBgNVBAoMB2V4YW1wbGUxFDASBgNVBAMMC2V4YW1wbGUuY29tMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDoVPW+x78bFdZZ/QFkwNPSimNtMik1vukX4FW/sBMozZtgPsBaEEXvTKguNAn54ReEr/x0iEgB8q8ml9pm/bfzPY3hKR4hBchhIWbfE6p75wL5tROBgsNR1my0atZJj9Q/OumhEWy4+3/rrrAN+9VJILom/MLy/+HpAYqiQ2oVbwIDAQABo1AwTjAdBgNVHQ4EFgQU6Jx//OWXkmm28irGVoFPl58IP8kwHwYDVR0jBBgwFoAU6Jx//OWXkmm28irGVoFPl58IP8kwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOBgQALsItfaZMdqPgGGNg7COEadWPapsai+9zT70pCsDQbPKKse22Nx4tyl21zDnGBtmk6x3tSL1b+DPwc8GUgL/XKszIVcHPNFHdsxiwP5CWQ7zeAaP9B5jJBCH5JWe1ciYbOpnyUyZrFyYS3TeArdfeA23u4ZPF5SjM9wOxXyF3AMw==</ds:X509Certificate>
		</ds:X509Data>
	  </ds:KeyInfo>
	</md:KeyDescriptor>
	<md:KeyDescriptor use="encryption">
	  <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
		<ds:X509Data>
		  <ds:X509Certificate>MIICTDCCAbWgAwIBAgIBADANBgkqhkiG9w0BAQ0FADBDMQswCQYDVQQGEwJmcjEMMAoGA1UECAwDSWRmMRAwDgYDVQQKDAdleGFtcGxlMRQwEgYDVQQDDAtleGFtcGxlLmNvbTAeFw0yMzAxMTExNDU0MjVaFw0yNDAxMTExNDU0MjVaMEMxCzAJBgNVBAYTAmZyMQwwCgYDVQQIDANJZGYxEDAOBgNVBAoMB2V4YW1wbGUxFDASBgNVBAMMC2V4YW1wbGUuY29tMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDoVPW+x78bFdZZ/QFkwNPSimNtMik1vukX4FW/sBMozZtgPsBaEEXvTKguNAn54ReEr/x0iEgB8q8ml9pm/bfzPY3hKR4hBchhIWbfE6p75wL5tROBgsNR1my0atZJj9Q/OumhEWy4+3/rrrAN+9VJILom/MLy/+HpAYqiQ2oVbwIDAQABo1AwTjAdBgNVHQ4EFgQU6Jx//OWXkmm28irGVoFPl58IP8kwHwYDVR0jBBgwFoAU6Jx//OWXkmm28irGVoFPl58IP8kwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOBgQALsItfaZMdqPgGGNg7COEadWPapsai+9zT70pCsDQbPKKse22Nx4tyl21zDnGBtmk6x3tSL1b+DPwc8GUgL/XKszIVcHPNFHdsxiwP5CWQ7zeAaP9B5jJBCH5JWe1ciYbOpnyUyZrFyYS3TeArdfeA23u4ZPF5SjM9wOxXyF3AMw==</ds:X509Certificate>
		</ds:X509Data>
	  </ds:KeyInfo>
	</md:KeyDescriptor>
	<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
	<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://example.com"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>
`
)

func testAccResourceExternalAuthSamlCreate() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_ExternalAuthSaml" {
  authentication_name = "testacc_ExternalAuthSaml"
  idp_metadata        = <<EOT
%s
EOT
  timeout             = 30
}
`, idpMetadataSAML)
}

func testAccResourceExternalAuthSamlUpdate() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_ExternalAuthSaml" {
  authentication_name = "testacc_ExternalAuthSaml"
  idp_metadata        = <<EOT
%s
EOT
  timeout             = 60
  description         = "testacc_ExternalAuthSaml description"
  certificate         = tls_self_signed_cert.example.cert_pem
  private_key         = tls_private_key.example.private_key_pem
}

resource "tls_private_key" "example" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P384"
}

resource "tls_self_signed_cert" "example" {
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
`, idpMetadataSAML)
}
