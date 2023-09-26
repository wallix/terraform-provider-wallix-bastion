# wallix-bastion_localpasswordpolicy Data Source

Get information on a localpasswordpolicy.

## Example Usage

```hcl
data "wallix-bastion_localpasswordpolicy" "default" {}
```

## Argument Reference

The following arguments are supported:

- **password_policy_name** (Optional, String)  
  The local password policy name.  
  Default to `default`.

## Attribute Reference

- **id** (String)  
  The configuration id.
- **allow_same_user_and_password** (Boolean)  
  Allow same username and password.
- **forbidden_passwords** (Set of String)  
  The list of forbidden passwords.
- **last_passwords_to_reject** (Number)  
  The number of last used passwords to reject.
- **max_auth_failures** (Number)  
  The maximum number of authentication failures allowed per user (0 = no limit).
- **password_expiration** (Number)  
  The number of days for password expiration (0 = never expires).
- **password_min_digit_chars** (Number)  
  The minimum number of digit chars in password.
- **password_min_length** (Number)  
  Minimum password length.
- **password_min_lower_chars** (Number)  
  The minimum number of lower case chars in password
- **password_min_special_chars** (Number)  
  The minimum number of special chars in password.
- **password_min_upper_chars** (Number)  
  The minimum number of upper case chars in password.
- **password_warning_days** (Number)  
  How many days the user should be warned about its password expiration (0 = no warning).
- **ssh_key_algos_allowed** (Set of String)  
  The list of SSH key algorithms allowed.
- **ssh_rsa_min_length** (Number)  
  The minimum RSA key length, in bits.
