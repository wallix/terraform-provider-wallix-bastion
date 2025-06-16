package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuthDomainSAML_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAuthDomainSAMLCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_authdomain_saml.testacc_AuthDomainSAML",
						"id"),
				),
			},
			{
				Config: testAccResourceAuthDomainSAMLUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_authdomain_saml.testacc_AuthDomainSAML",
				ImportState:   true,
				ImportStateId: "testacc.AuthDomainSAML-u",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceAuthDomainSAMLCreate() string {
	return `
resource "wallix-bastion_authdomain_saml" "testacc_AuthDomainSAML" {
  domain_name          = "testacc.AuthDomainSAML"
  auth_domain_name     = "test4.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_AuthDomainSAML.authentication_name]
  default_email_domain = "test4.com"
  default_language     = "fr"
  label                = "SAML"
}
resource "wallix-bastion_externalauth_saml" "testacc_AuthDomainSAML" {
  authentication_name = "testacc_AuthDomainSAML"
  idp_metadata        = local.idp_metadata
  timeout             = 120
  claim_customization {
    username    = "username"
    displayname = "displayname"
    email       = "email"
    group       = "group"
  }
}
` + testAccResourceAuthDomainSAMLIdpMetadata()
}

func testAccResourceAuthDomainSAMLUpdate() string {
	return `
resource "wallix-bastion_authdomain_saml" "testacc_AuthDomainSAML" {
  domain_name          = "testacc.AuthDomainSAML-u"
  auth_domain_name     = "test4.com"
  external_auths       = [wallix-bastion_externalauth_saml.testacc_AuthDomainSAML.authentication_name]
  default_email_domain = "test4.com"
  default_language     = "fr"
  label                = "SAML test4.com"
  description          = "SAML test4.com"
  force_authn          = true
  is_default           = true
}
resource "wallix-bastion_externalauth_saml" "testacc_AuthDomainSAML" {
  authentication_name = "testacc_AuthDomainSAML"
  idp_metadata        = local.idp_metadata
  timeout             = 120
  claim_customization {
    username    = "username"
    displayname = "displayname"
    email       = "email"
    group       = "group"
  }
}
` + testAccResourceAuthDomainSAMLIdpMetadata()
}

//nolint:lll
func testAccResourceAuthDomainSAMLIdpMetadata() string {
	return `
locals {
	idp_metadata = <<EOF
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
