package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// The claim_customization block requires API v3.12+; skip on older versions. Default (unset) is v3.12+.
func TestAccDataSourceAuthDomainSAML_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resource.Test(t, resource.TestCase{
			PreCheck:                  func() { testAccPreCheck(t) },
			Providers:                 testAccProviders,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					// Create the resource to be fetched by the datasource.
					Config: testAccDataSourceAuthDomainSAMLConfigCreate(),
				},
				{
					// Validate that the datasource correctly retrieves the resource.
					Config: testAccDataSourceAuthDomainSAMLConfigData(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML",
							"domain_name", "testacc-domain-saml-ds"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML",
							"auth_domain_name", "test4-ds.com"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML",
							"default_email_domain", "test4-ds.com"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML",
							"default_language", "fr"),
						resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML",
							"label", "SAML"),
					),
				},
			},
		})
	}
}

// Resource creation configuration.
func testAccDataSourceAuthDomainSAMLConfigCreate() string {
	return `
resource "wallix-bastion_authdomain_saml" "testacc_dataAuthDomainSAML_ds" {
  domain_name          = "testacc-domain-saml-ds"
  auth_domain_name     = "test4-ds.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_dataAuthDomainSAML_ds.authentication_name]
  default_email_domain = "test4-ds.com"
  default_language     = "fr"
  label                = "SAML"
}
resource "wallix-bastion_externalauth_saml" "testacc_dataAuthDomainSAML_ds" {
  authentication_name = "testacc_dataAuthDomainSAML_ds"
  idp_metadata        = local.idp_metadata_saml_ds
  timeout             = 120
  claim_customization {
    username    = "username"
    displayname = "displayname"
    email       = "email"
    group       = "group"
  }
}
` + testAccDataSourceAuthDomainSAMLIdpMetadata()
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainSAMLConfigData() string {
	return `
resource "wallix-bastion_authdomain_saml" "testacc_dataAuthDomainSAML_ds" {
  domain_name          = "testacc-domain-saml-ds"
  auth_domain_name     = "test4-ds.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_dataAuthDomainSAML_ds.authentication_name]
  default_email_domain = "test4-ds.com"
  default_language     = "fr"
  label                = "SAML"
}
resource "wallix-bastion_externalauth_saml" "testacc_dataAuthDomainSAML_ds" {
  authentication_name = "testacc_dataAuthDomainSAML_ds"
  idp_metadata        = local.idp_metadata_saml_ds
  timeout             = 120
  claim_customization {
    username    = "username"
    displayname = "displayname"
    email       = "email"
    group       = "group"
  }
}

data "wallix-bastion_authdomain_saml" "testacc_dataAuthDomainSAML" {
  domain_name = wallix-bastion_authdomain_saml.testacc_dataAuthDomainSAML_ds.domain_name
}
` + testAccDataSourceAuthDomainSAMLIdpMetadata()
}

//nolint:lll
func testAccDataSourceAuthDomainSAMLIdpMetadata() string {
	return `
locals {
	idp_metadata_saml_ds = <<EOF
<EntityDescriptor ID="_c066524f-ba36-49d5-9dfa-ae14e13c1392" entityID="https://idp.identityserver" validUntil="2022-07-20T09:48:54Z" cacheDuration="PT15M" xmlns="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
    <IDPSSODescriptor WantAuthnRequestsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://idp.identityserver/saml/sso" />
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://idp.identityserver/saml/sso" />
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="https://idp.identityserver/saml/sso" />

        <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://idp.identityserver/saml/slo" />
        <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://idp.identityserver/saml/slo" />
        <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="https://idp.identityserver/saml/slo" />

        <ArtifactResolutionService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="https://idp.identityserver/saml/ars" index="0" />

        <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</NameIDFormat>
        <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</NameIDFormat>
        <NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</NameIDFormat>
        <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>

        <KeyDescriptor use="signing">
            <KeyInfo
                xmlns="http://www.w3.org/2000/09/xmldsig#">
                <X509Data>
                    <X509Certificate>IDP_PUBLIC_SIGNING_CERTIFICATE_USED_FOR_SIGNING_RESPONSES</X509Certificate>
                </X509Data>
            </KeyInfo>
        </KeyDescriptor>
    </IDPSSODescriptor>

    <Organization>
        <OrganizationName xml:lang="en-GB">Example</OrganizationName>
        <OrganizationDisplayName xml:lang="en-GB">Example Org</OrganizationDisplayName>
        <OrganizationURL xml:lang="en-GB">https://example.com/</OrganizationURL>
    </Organization>

    <ContactPerson contactType="technical">
        <Company>Example</Company>
        <GivenName>bob</GivenName>
        <SurName>smith</SurName>
        <EmailAddress>bob@example.com</EmailAddress>
    </ContactPerson>

</EntityDescriptor>
EOF
}
`
}
