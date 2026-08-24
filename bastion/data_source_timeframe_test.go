package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTimeframe_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceTimeframeConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceTimeframeConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_timeframe.testacc_dataTimeframe",
						"timeframe_name", "testacc_dataTimeframe"),
					resource.TestCheckResourceAttr("data.wallix-bastion_timeframe.testacc_dataTimeframe",
						"description", "testacc data Timeframe"),
					resource.TestCheckResourceAttr("data.wallix-bastion_timeframe.testacc_dataTimeframe",
						"is_overtimable", "true"),
					resource.TestCheckResourceAttr("data.wallix-bastion_timeframe.testacc_dataTimeframe",
						"periods.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_timeframe.testacc_dataTimeframe",
						"periods.*", map[string]string{
							"start_date": "2020-01-01",
							"end_date":   "2020-02-02",
							"start_time": "08:00",
							"end_time":   "12:00",
						}),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceTimeframeConfigCreate() string {
	return `
resource "wallix-bastion_timeframe" "testacc_dataTimeframe" {
  timeframe_name = "testacc_dataTimeframe"
  description    = "testacc data Timeframe"
  is_overtimable = true
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceTimeframeConfigData() string {
	return `
resource "wallix-bastion_timeframe" "testacc_dataTimeframe" {
  timeframe_name = "testacc_dataTimeframe"
  description    = "testacc data Timeframe"
  is_overtimable = true
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}

data "wallix-bastion_timeframe" "testacc_dataTimeframe" {
  timeframe_name = wallix-bastion_timeframe.testacc_dataTimeframe.timeframe_name
}
`
}
