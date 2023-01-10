# wallix-bastion_timeframe Resource

Provides a timeframe resource.

## Example Usage

```hcl
# Configure a timeframe
resource "wallix-bastion_timeframe" "demo" {
  timeframe_name = "demo"
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}
```

## Argument Reference

The following arguments are supported:

- **timeframe_name** (Required, String, Forces new resource)  
  The timeframe name.
- **description** (Optional, String)  
  The timeframe description.
- **is_overtimable** (Optional, Boolean)  
  Do not close sessions at the end of the time period.
- **periods** (Optional, Set of Block)  
  The timeframe periods.  
  Can be specified multiple times for each period to declare.
  - **start_date** (Required, String)  
  The period start date.  
  Must respect the format `yyyy-mm-dd`.
  - **end_date** (Required, String)  
  The period end date.  
  Must respect the format `yyyy-mm-dd`.
  - **start_time** (Required, String)  
  The period start time.  
  Must respect the format `hh:mm`.
  - **end_time** (Required, String)  
  The period end time.  
  Must respect the format `hh:mm`.
  - **week_days** (Required, List of String)  
  The period week days.  
  Elements need to be `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday` or `sunday`.

## Attribute Reference

- **id** (String)  
  ID of resource = `timeframe_name`

## Import

Timeframe can be imported using an id made up of `<timeframe_name>`, e.g.

```shell
terraform import wallix-bastion_timeframe.demo demo
```
