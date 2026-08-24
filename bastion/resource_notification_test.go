package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNotification_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNotificationCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_notification.testacc_Notification",
						"id"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_notification.testacc_Notification",
						"events.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"wallix-bastion_notification.testacc_Notification",
						"events.*", "password_expired"),
					resource.TestCheckTypeSetElemAttr(
						"wallix-bastion_notification.testacc_Notification",
						"events.*", "raid_error"),
				),
			},
			{
				// Replaces the events set entirely (wrong_fingerprint instead of
				// password_expired/raid_error) - exercises the force=true full-replace PUT,
				// since without it the API would union this into the existing set instead.
				Config: testAccResourceNotificationUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_notification.testacc_Notification",
						"description", "testacc updated description"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_notification.testacc_Notification",
						"events.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"wallix-bastion_notification.testacc_Notification",
						"events.*", "wrong_fingerprint"),
				),
			},
			{
				ResourceName:  "wallix-bastion_notification.testacc_Notification",
				ImportState:   true,
				ImportStateId: "testacc_Notification",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceNotificationCreate() string {
	return `
resource "wallix-bastion_notification" "testacc_Notification" {
  notification_name = "testacc_Notification"
  enabled           = true
  type              = "email"
  destination       = "testacc@example.com"
  language          = "en"
  events            = ["password_expired", "raid_error"]
}
`
}

func testAccResourceNotificationUpdate() string {
	return `
resource "wallix-bastion_notification" "testacc_Notification" {
  notification_name = "testacc_Notification"
  description       = "testacc updated description"
  enabled           = true
  type              = "email"
  destination       = "testacc@example.com"
  language          = "en"
  events            = ["wrong_fingerprint"]
}
`
}
