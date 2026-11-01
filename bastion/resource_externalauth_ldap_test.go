package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceExternalAuthLDAP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthLDAPCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAP",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthLDAPUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAP",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthLDAP",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthLDAPCreate() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_ExternalAuthLDAP" {
  authentication_name = "testacc_ExternalAuthLDAP"
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

func testAccResourceExternalAuthLDAPUpdate() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_ExternalAuthLDAP" {
  authentication_name = "testacc_ExternalAuthLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_active_directory = true
  login               = "svc1"
  password            = "aPassword"
  description         = "testacc ExternalAuthLDAP"
  is_protected_user   = true
}
`
}

// The API reads ca_certificate back as the subject of the CA; each apply must still plan empty (#195).
func TestAccResourceExternalAuthLDAP_caCertificate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			tvTLSProviderName: {Source: tvTLSProviderSource},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthLDAPCACertificate("testacc ExternalAuthLDAP CA"),
				Check: resource.TestCheckResourceAttrPair(
					"wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAPCA", "ca_certificate",
					"tls_self_signed_cert.testacc_ExternalAuthLDAPCA", "cert_pem"),
			},
			{
				Config: testAccResourceExternalAuthLDAPCACertificate("testacc ExternalAuthLDAP CA2"),
				Check: resource.TestCheckResourceAttrPair(
					"wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAPCA", "ca_certificate",
					"tls_self_signed_cert.testacc_ExternalAuthLDAPCA", "cert_pem"),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAPCA",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthLDAPCA",
				// Without a prior PEM, the import stores the subject the API returns.
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					want := "/O=testacc/CN=testacc ExternalAuthLDAP CA2"
					if got := states[0].Attributes["ca_certificate"]; got != want {
						return fmt.Errorf("ca_certificate = %q after import, want %q", got, want)
					}

					return nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// 0.17.0 stored the subject in the state: one update sends the PEM again, then plans are empty.
func TestAccResourceExternalAuthLDAP_upgradeCACertificate(t *testing.T) {
	t.Setenv(resource.EnvTfAccProviderNamespace, "wallix")
	tls := map[string]resource.ExternalProvider{tvTLSProviderName: {Source: tvTLSProviderSource}}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"wallix-bastion":  {Source: "wallix/wallix-bastion", VersionConstraint: "0.17.0"},
					tvTLSProviderName: {Source: tvTLSProviderSource},
				},
				Config:             testAccResourceExternalAuthLDAPCACertificate("testacc ExternalAuthLDAP CA"),
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				ExternalProviders:        tls,
				Config: testAccResourceExternalAuthLDAPUpgradeSource +
					testAccResourceExternalAuthLDAPCACertificate("testacc ExternalAuthLDAP CA"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				ExternalProviders:        tls,
				Config: testAccResourceExternalAuthLDAPUpgradeSource +
					testAccResourceExternalAuthLDAPCACertificate("testacc ExternalAuthLDAP CA"),
				Check: resource.TestCheckResourceAttrPair(
					"wallix-bastion_externalauth_ldap.testacc_ExternalAuthLDAPCA", "ca_certificate",
					"tls_self_signed_cert.testacc_ExternalAuthLDAPCA", "cert_pem"),
			},
		},
	})
}

const testAccResourceExternalAuthLDAPUpgradeSource = `
terraform {
  required_providers {
    wallix-bastion = {
      source = "wallix/wallix-bastion"
    }
  }
}
`

func testAccResourceExternalAuthLDAPCACertificate(commonName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "testacc_ExternalAuthLDAPCA" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "tls_self_signed_cert" "testacc_ExternalAuthLDAPCA" {
  private_key_pem = tls_private_key.testacc_ExternalAuthLDAPCA.private_key_pem
  subject {
    common_name  = %q
    organization = "testacc"
  }
  validity_period_hours = 24
  is_ca_certificate     = true
  allowed_uses          = ["cert_signing", "crl_signing"]
}

resource "wallix-bastion_externalauth_ldap" "testacc_ExternalAuthLDAPCA" {
  authentication_name = "testacc_ExternalAuthLDAPCA"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
  ca_certificate      = tls_self_signed_cert.testacc_ExternalAuthLDAPCA.cert_pem
}
`, commonName)
}
