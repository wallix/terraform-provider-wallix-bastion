package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNotification_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceNotificationConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceNotificationConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"notification_name", "testacc_dataNotification"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"description", "testacc data notification"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"enabled", "true"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"type", "email"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"destination", "testacc@example.com"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"language", "en"),
					resource.TestCheckResourceAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"events.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"events.*", "password_expired"),
					resource.TestCheckTypeSetElemAttr("data.wallix-bastion_notification.testacc_dataNotification",
						"events.*", "raid_error"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceNotificationConfigCreate() string {
	return `
resource "wallix-bastion_notification" "testacc_dataNotification" {
  notification_name = "testacc_dataNotification"
  description        = "testacc data notification"
  enabled            = true
  type               = "email"
  destination        = "testacc@example.com"
  language           = "en"
  events             = ["password_expired", "raid_error"]
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceNotificationConfigData() string {
	return `
resource "wallix-bastion_notification" "testacc_dataNotification" {
  notification_name = "testacc_dataNotification"
  description        = "testacc data notification"
  enabled            = true
  type               = "email"
  destination        = "testacc@example.com"
  language           = "en"
  events             = ["password_expired", "raid_error"]
}

data "wallix-bastion_notification" "testacc_dataNotification" {
  notification_name = wallix-bastion_notification.testacc_dataNotification.notification_name
}
`
}
