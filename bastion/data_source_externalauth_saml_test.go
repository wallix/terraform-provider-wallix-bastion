// nolint: lll,nolintlint
package bastion_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

func TestAccDataSourceExternalAuthSaml_basic38(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == bastion.VersionWallixAPI38 {
		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			ProviderFactories:         testAccProviderFactories,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceExternalAuthSaml38ConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceExternalAuthSaml38ConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"authentication_name", "testacc_dataExternalAuthSaml"),
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"timeout", "30"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"idp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"saml_request_url"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"saml_request_method"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_metadata"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_assertion_consumer_service"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_single_logout_service"),
					),
				},
			},
		})
	}
}

// WALLIX_BASTION_API_VERSION defaults to v3.12 when unset (see provider.go's EnvDefaultFunc), so
// an empty env var must be treated the same as an explicit v3.12 here.
func TestAccDataSourceExternalAuthSaml_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || v == bastion.VersionWallixAPI312 {
		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			ProviderFactories:         testAccProviderFactories,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceExternalAuthSamlConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceExternalAuthSamlConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"authentication_name", "testacc_dataExternalAuthSaml"),
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"timeout", "30"),
						resource.TestCheckResourceAttr(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"claim_customization.0.username", "email"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"idp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"saml_request_url"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"saml_request_method"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_metadata"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_entity_id"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_assertion_consumer_service"),
						resource.TestCheckResourceAttrSet(
							"data.wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml",
							"sp_single_logout_service"),
					),
				},
			},
		})
	}
}

const (
	//nolint: lll
	dataSourceIdpMetadataSAML = `<?xml version="1.0"?>
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

// Resource creation configuration (WAB v3.8, no claim_customization support).
func testAccDataSourceExternalAuthSaml38ConfigCreate() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = "testacc_dataExternalAuthSaml"
  idp_metadata         = <<EOT
%s
EOT
  timeout              = 30
}
`, dataSourceIdpMetadataSAML)
}

// Datasource configuration to retrieve the created resource (WAB v3.8).
func testAccDataSourceExternalAuthSaml38ConfigData() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = "testacc_dataExternalAuthSaml"
  idp_metadata         = <<EOT
%s
EOT
  timeout              = 30
}

data "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml.authentication_name
}
`, dataSourceIdpMetadataSAML)
}

// Resource creation configuration (WAB > v3.8, with claim_customization support).
func testAccDataSourceExternalAuthSamlConfigCreate() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = "testacc_dataExternalAuthSaml"
  idp_metadata         = <<EOT
%s
EOT
  timeout              = 30
  claim_customization {
    username = "email"
  }
}
`, dataSourceIdpMetadataSAML)
}

// Datasource configuration to retrieve the created resource (WAB > v3.8).
func testAccDataSourceExternalAuthSamlConfigData() string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = "testacc_dataExternalAuthSaml"
  idp_metadata         = <<EOT
%s
EOT
  timeout              = 30
  claim_customization {
    username = "email"
  }
}

data "wallix-bastion_externalauth_saml" "testacc_dataExternalAuthSaml" {
  authentication_name = wallix-bastion_externalauth_saml.testacc_dataExternalAuthSaml.authentication_name
}
`, dataSourceIdpMetadataSAML)
}
