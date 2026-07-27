package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDomainAzureAD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAuthDomainAzureADConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAuthDomainAzureADConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"domain_name", "testacc-domain-azuread-ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"auth_domain_name", "testacc-auth-domain-azuread-ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"client_id", "00000000-0000-0000-0000-000000000001"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"entity_id", "https://sts.windows.net/00000000-0000-0000-0000-000000000000/"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"label", "AzureAD"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"default_language", "en"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD",
						"default_email_domain", "example.com"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAuthDomainAzureADConfigCreate() string {
	return `
resource "wallix-bastion_externalauth_saml" "testacc_dataAuthDomainAzureAD_ds" {
  authentication_name = "testacc_dataAuthDomainAzureAD_ds"
  idp_metadata         = local.idp_metadata_azuread_ds
  timeout              = 120
}

resource "wallix-bastion_authdomain_azuread" "testacc_dataAuthDomainAzureAD_ds" {
  domain_name          = "testacc-domain-azuread-ds"
  auth_domain_name     = "testacc-auth-domain-azuread-ds"
  client_id            = "00000000-0000-0000-0000-000000000001"
  entity_id            = "https://sts.windows.net/00000000-0000-0000-0000-000000000000/"
  label                = "AzureAD"
  default_language     = "en"
  default_email_domain = "example.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_dataAuthDomainAzureAD_ds.authentication_name]
}
` + testAccDataSourceAuthDomainAzureADIdpMetadata()
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainAzureADConfigData() string {
	return `
resource "wallix-bastion_externalauth_saml" "testacc_dataAuthDomainAzureAD_ds" {
  authentication_name = "testacc_dataAuthDomainAzureAD_ds"
  idp_metadata         = local.idp_metadata_azuread_ds
  timeout              = 120
}

resource "wallix-bastion_authdomain_azuread" "testacc_dataAuthDomainAzureAD_ds" {
  domain_name          = "testacc-domain-azuread-ds"
  auth_domain_name     = "testacc-auth-domain-azuread-ds"
  client_id            = "00000000-0000-0000-0000-000000000001"
  entity_id            = "https://sts.windows.net/00000000-0000-0000-0000-000000000000/"
  label                = "AzureAD"
  default_language     = "en"
  default_email_domain = "example.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_dataAuthDomainAzureAD_ds.authentication_name]
}

data "wallix-bastion_authdomain_azuread" "testacc_dataAuthDomainAzureAD" {
  domain_name = wallix-bastion_authdomain_azuread.testacc_dataAuthDomainAzureAD_ds.domain_name
}
` + testAccDataSourceAuthDomainAzureADIdpMetadata()
}

//nolint:lll
func testAccDataSourceAuthDomainAzureADIdpMetadata() string {
	return `
locals {
	idp_metadata_azuread_ds = <<EOF
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
