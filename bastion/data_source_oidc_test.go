package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOIDC_basic(t *testing.T) {
	testAccSkipBeforeOIDC(t)
	auth := "data.wallix-bastion_externalauth_oidc.testacc_dataOIDC"
	domain := "data.wallix-bastion_authdomain_oidc.testacc_dataOIDC"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOIDCConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(auth, "client_id", "testacc"),
					resource.TestCheckResourceAttr(auth, "timeout", "30"),
					resource.TestCheckResourceAttr(auth, "claim_customization.username", "preferred_username"),
					resource.TestCheckResourceAttr(auth, "oidc_advanced_attributes.jwks_uri", "https://idp.example.com/jwks"),
					resource.TestCheckResourceAttrSet(auth, "redirect_uri"),
					resource.TestCheckResourceAttr(domain, "external_auths.0", "testacc_dataOIDC"),
					resource.TestCheckResourceAttr(domain, "label", "testacc OIDC"),
					resource.TestCheckResourceAttr(domain, "is_default", "false"),
				),
			},
		},
	})
}

const testAccDataSourceOIDCConfig = `
resource "wallix-bastion_externalauth_oidc" "testacc_dataOIDC" {
  authentication_name = "testacc_dataOIDC"
  timeout             = 30
  client_id           = "testacc"
  claim_customization {
    username = "preferred_username"
  }
  oidc_advanced_attributes {
    issuer                                = "https://idp.example.com"
    scope                                 = ["openid"]
    authorization_endpoint                = "https://idp.example.com/authorize"
    token_endpoint                        = "https://idp.example.com/token"
    jwks_uri                              = "https://idp.example.com/jwks"
    id_token_signing_alg_values_supported = ["RS256"]
  }
}
resource "wallix-bastion_authdomain_oidc" "testacc_dataOIDC" {
  domain_name          = "testacc.dataOIDC"
  auth_domain_name     = "testacc-data-oidc.example.com"
  external_auths       = [wallix-bastion_externalauth_oidc.testacc_dataOIDC.authentication_name]
  default_email_domain = "example.com"
  default_language     = "en"
  label                = "testacc OIDC"
}
data "wallix-bastion_externalauth_oidc" "testacc_dataOIDC" {
  authentication_name = wallix-bastion_externalauth_oidc.testacc_dataOIDC.authentication_name
}
data "wallix-bastion_authdomain_oidc" "testacc_dataOIDC" {
  domain_name = wallix-bastion_authdomain_oidc.testacc_dataOIDC.domain_name
}
`
