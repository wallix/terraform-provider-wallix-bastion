package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTimeframe_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTimeframeCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"id"),
				),
			},
			{
				Config: testAccResourceTimeframeUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"description", "testacc Timeframe"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"periods.#", "2"),
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_timeframe.testacc_Timeframe2",
						"id"),
				),
			},
			{
				ResourceName:  "wallix-bastion_timeframe.testacc_Timeframe",
				ImportState:   true,
				ImportStateId: "testacc_Timeframe",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceTimeframeCreate() string {
	return `
resource "wallix-bastion_timeframe" "testacc_Timeframe" {
  timeframe_name = "testacc_Timeframe"
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}
resource "wallix-bastion_timeframe" "testacc_Timeframe2" {
  timeframe_name = "testacc_Timeframe2"
}
`
}

func testAccResourceTimeframeUpdate() string {
	return `
resource "wallix-bastion_timeframe" "testacc_Timeframe" {
  timeframe_name = "testacc_Timeframe"
  description    = "testacc Timeframe"
  is_overtimable = true
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
  periods {
    start_date = "2020-02-01"
    end_date   = "2020-03-02"
    start_time = "10:00"
    end_time   = "16:00"
    week_days  = ["monday", "friday"]
  }
}
resource "wallix-bastion_timeframe" "testacc_Timeframe2" {
  timeframe_name = "testacc_Timeframe2"
}
`
}
