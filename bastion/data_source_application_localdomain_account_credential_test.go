package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceApplicationLocalDomainAccountCred_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceApplicationLocalDomainAccountCredConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceApplicationLocalDomainAccountCredConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_application_localdomain_account_credential.testacc_dataAppLocalDomAccountCred",
						"type", "password"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account_credential.testacc_dataAppLocalDomAccountCred",
						"password",
						"random_password.testacc_dataAppLocalDomAccountCred", "result"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account_credential.testacc_dataAppLocalDomAccountCred",
						"application_id",
						"wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds", "id"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account_credential.testacc_dataAppLocalDomAccountCred",
						"domain_id",
						"wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds", "id"),
					resource.TestCheckResourceAttrPair(
						"data.wallix-bastion_application_localdomain_account_credential.testacc_dataAppLocalDomAccountCred",
						"account_id",
						"wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccountCred_ds", "id"),
				),
			},
		},
	})
}

// Resource creation configuration.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainAccountCredConfigCreate() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDomAccountCred_ds" {
  device_name = "testacc_dataAppLocalDomAccountCred_ds"
  host        = "192.168.100.16"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDomAccountCred_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.id
  service_name      = "testacc_dataAppLocalDomAccountCred_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDomAccountCred_ds" {
  cluster_name = "testacc_dataAppLocalDomAccountCred_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccountCred_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDomAccountCred_ds" {
  application_name  = "testacc_dataAppLocalDomAccountCred_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccountCred_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDomAccountCred_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_name    = "testacc_dataAppLocalDomAccountCred_ds"
}

resource "wallix-bastion_application_localdomain_account" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds.id
  account_name   = "testacc_dataAppLocalDomAccountCred_ds_admin"
  account_login  = "admin"
}

resource "wallix-bastion_application_localdomain_account_credential" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds.id
  account_id     = wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccountCred_ds.id
  type           = "password"
  password       = random_password.testacc_dataAppLocalDomAccountCred.result
}

resource "random_password" "testacc_dataAppLocalDomAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
`
}

// Datasource configuration to retrieve the created resource.
// nolint: lll, nolintlint
func testAccDataSourceApplicationLocalDomainAccountCredConfigData() string {
	return `
resource "wallix-bastion_device" "testacc_dataAppLocalDomAccountCred_ds" {
  device_name = "testacc_dataAppLocalDomAccountCred_ds"
  host        = "192.168.100.16"
}

resource "wallix-bastion_device_service" "testacc_dataAppLocalDomAccountCred_ds" {
  device_id         = wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.id
  service_name      = "testacc_dataAppLocalDomAccountCred_ds"
  connection_policy = "RDP"
  port              = 22
  protocol          = "RDP"
  subprotocols      = ["RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT"]
}

resource "wallix-bastion_cluster" "testacc_dataAppLocalDomAccountCred_ds" {
  cluster_name = "testacc_dataAppLocalDomAccountCred_ds"
  interactive_logins = [
    "${wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccountCred_ds.service_name}",
  ]
}

resource "wallix-bastion_application" "testacc_dataAppLocalDomAccountCred_ds" {
  application_name  = "testacc_dataAppLocalDomAccountCred_ds_application"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@${wallix-bastion_device.testacc_dataAppLocalDomAccountCred_ds.device_name}:${wallix-bastion_device_service.testacc_dataAppLocalDomAccountCred_ds.service_name}"
    program     = "application_path"
    working_dir = "directory"
  }
  target = wallix-bastion_cluster.testacc_dataAppLocalDomAccountCred_ds.cluster_name
}

resource "wallix-bastion_application_localdomain" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_name    = "testacc_dataAppLocalDomAccountCred_ds"
}

resource "wallix-bastion_application_localdomain_account" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds.id
  account_name   = "testacc_dataAppLocalDomAccountCred_ds_admin"
  account_login  = "admin"
}

resource "wallix-bastion_application_localdomain_account_credential" "testacc_dataAppLocalDomAccountCred_ds" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds.id
  account_id     = wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccountCred_ds.id
  type           = "password"
  password       = random_password.testacc_dataAppLocalDomAccountCred.result
}

resource "random_password" "testacc_dataAppLocalDomAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}

data "wallix-bastion_application_localdomain_account_credential" "testacc_dataAppLocalDomAccountCred" {
  application_id = wallix-bastion_application.testacc_dataAppLocalDomAccountCred_ds.id
  domain_id      = wallix-bastion_application_localdomain.testacc_dataAppLocalDomAccountCred_ds.id
  account_id     = wallix-bastion_application_localdomain_account.testacc_dataAppLocalDomAccountCred_ds.id
  type           = "password"
}
`
}
