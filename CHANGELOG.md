<!-- markdownlint-disable-file MD013 MD041 -->
# changelog

* resource/**connection_policy**: `authentication_methods` is now unordered

## 0.4.1 (October 18, 2021)

ENHANCEMENTS:

* [docs] reformat arguments/attributes, add example usage & minor fix

BUG FIXES:

* fix the potential double slash in url when calling Wallix API
* fix missing sensitive options for few arguments
* resource/**wallix-bastion_application_localdomain**, **wallix-bastion_device_localdomain**, **wallix-bastion_domain**: fix arguments requirement
* resource/**wallix-bastion_application**: fix panic with `global_domains`
* resource/**wallix-bastion_profile**: fix `default_target_group` is required in `target_groups_limitation` block
* resource/**wallix-bastion_domain**: fix `passphrase` can't refresh
* resource/**wallix-bastion_device_localdomain**: fix `passphrase` can't refresh

## 0.4.0 (April 9, 2021)

FEATURES:

* add `wallix-bastion_domain` data source

## 0.3.3 (April 6, 2021)

BUG FIXES:

* fix `global_domains` argument can be an attribute in `wallix-bastion_device_service` resource

## 0.3.2 (April 1, 2021)

BUG FIXES:

* fix `device`/`service` or `application` needed with `domain_type`="global" on `session_accounts` in `wallix-bastion_targetgroup` resource
* fix `resources` argument can be an attribute in `wallix-bastion_domain_account` resource

## 0.3.1 (March 30, 2021)

BUG FIXES:

* fix import user resource

## 0.3.0 (March 19, 2021)

FEATURES:

* add `wallix-bastion_application` resource
* add `wallix-bastion_application_localdomain` resource
* add `wallix-bastion_application_localdomain_account` resource
* add `wallix-bastion_checkout_policy` resource
* add `wallix-bastion_cluster` resource
* add `wallix-bastion_connection_policy` resource
* add `wallix-bastion_externalauth_kerberos` resource
* add `wallix-bastion_externalauth_radius` resource
* add `wallix-bastion_externalauth_tacacs` resource
* add `wallix-bastion_profile` resource
* add `wallix-bastion_timeframe` resource

## 0.2.0 (March 5, 2021)

FEATURES:

* add `wallix-bastion_authorization` resource
* add `wallix-bastion_device`resource
* add `wallix-bastion_device_localdomain` resource
* add `wallix-bastion_device_localdomain_account` resource
* add `wallix-bastion_device_localdomain_account_credential` resource
* add `wallix-bastion_device_service` resource
* add `wallix-bastion_domain` resource
* add `wallix-bastion_domain_account` resource
* add `wallix-bastion_domain_account_credential` resource
* add `wallix-bastion_ldapdomain` resource
* add `wallix-bastion_ldapmapping` resource
* add `wallix-bastion_targetgroup` resource

ENHANCEMENTS:

* remove Forcenew on `authentication_name` in `wallix-bastion_externalauth_ldap` resource, it's not necessary

BUG FIXES:

* typo in errors displayed
* remove log to debug in http request (possible secret could appear)
* `timeframes` and `restrictions` aren't ordered in `wallix-bastion_usegroup` resource
* do not reactivate `force_change_pwd` after creation and the password has changed in `wallix-bastion_user` resource

## 0.1.0 (February 9, 2021)

First release
