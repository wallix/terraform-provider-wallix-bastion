# wallix-bastion_ldapdomain Resource

Provides a ldapdomain resource.

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required, Forces new resource)(`String`) The domain name.
* `ldap_domain_name` - (Required)(`String`) The LDAP domain name.
* `external_ldaps` - (Required)(`ListOfString`) The LDAP external authentications.
* `default_language` - (Required)(`String`) The default language. Need to be 'de', 'en', 'es', 'fr' or 'ru'.
* `default_email_domain` - (Required)(`String`) The default email domain.
* `secondary_auth` - (Optional)(`ListOfString`) The secondary authentications methods for the LDAP domain.
* `description` - (Optional)(`String`) The domain description.
* `check_x509_san_email` - (Optional)(`Bool`) Match the X509v3 SAN email.
* `display_name_attribute` - (Optional)(`String`) The display name attribute.
* `email_attribute` - (Optional)(`String`) The email attribute.
* `group_attribute` - (Optional)(`String`) The group attribute.
* `is_default` - (Optional)(`Bool`) The domain is used by default.
* `language_attribute` - (Optional)(`String`) The language attribute.
* `san_domain_name` - (Optional)(`String`) The domain name to match SAN email (only for AD server).
* `x509_condition` - (Optional)(`String`) Condition to match a LDAP domain with the X509 certificate variables (only for LDAP server).
* `x509_search_filter` - (Optional)(`String`) LDAP search filter for X509 authentication (only for LDAP server).

## Attribute Reference

* `id` - (`String`) = `domain_name`

## Import

Ldapdomain can be imported using an id made up of `<domain_name>`, e.g.

```
$ terraform import wallix-bastion_ldapdomain.example_com example.com
```
