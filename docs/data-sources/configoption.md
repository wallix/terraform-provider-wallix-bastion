# wallix-bastion_configoption Data Source

Get information on a configoption.

## Example Usage

```hcl
data "wallix-bastion_configoption" "global" {
  config_id = "wabsshkeys"
  options_list = [
    "signature_type",
  ]
  lifecycle {
    postcondition {
      condition     = jsondecode(self.options.0).value == "rsa-sha2-512"
      error_message = "wabsshkeys.signature_type is NOT rsa-sha2-512."
    }
  }
}
```

## Argument Reference

The following arguments are supported:

- **config_id** (Required, String)  
  Name or id of configuration.
- **options_list** (Optional, Set of String)  
  List of options to return, which is a list of sections or option names
  (with a section, all sub-options are returned).  
  Example: `["section1","sect2.field1","sect2.field2"]`

## Attribute Reference

- **id** (String)  
  The configuration id.
- **config_name** (String)  
  The configuration internal name.
- **name** (String)  
  The configuration name, for display.
- **date** (String)  
  Date of last change in the configuration file,
  empty string if this configuration file has never been changed.
- **options** (List of String)  
  List of sections and options in the sections.  
  Each string is a JSON to be decode.
