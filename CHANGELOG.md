# changelog

## 0.14.1 (Decembner 13, 2024)

FEATURES:

* **datasource/wallix-bastion_authdomain_ad**: added the datasource to retrieve an existing authdomain
* **resource/wallix-bastion_domain_account_credential**: added credential propagation to AD upon creation.

BUG FIXES:

* **resource/wallix-bastion_externalauth_kerberos**: deprecate `login_attribute` argument (it produces Bad Request with API v3.12)
* **provider_test**: Added the user environment variable presence test for acceptance tests.

## 0.14.0 (November 08, 2024)

BREAKING CHANGES:

* remove compatibility with API version 3.3 and 3.6
* remove resource `wallix-bastion_ldapdomain`
* remove resource `wallix-bastion_ldapmapping`
* default provider api_version argument is now `v3.8`
* user statement is now mandatory

FEATURES:

* add compatibility with API version 3.12

ENHANCEMENTS:

* **resource/wallix-bastion_application**:
  * add `category`, `application_url`, `browser`, `browser_version` arguments to be able to add `jumphost` application (not tested)
  * `paths` and `target` is now only required when `category` = `standard`
* **resource/wallix-bastion_connection_policy**: add `type` argument with default value as `protocol` value
* **resource/wallix-bastion_externalauth_saml**: add `claim_customization` block argument

## 0.13.0 (March 08, 2024)

* build(deps): bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 by @dependabot in https://github.com/wallix/terraform-provider-wallix-bastion/pull/13
* added http basic authentication by @moulip in https://github.com/wallix/terraform-provider-wallix-bastion/pull/15

## 0.12.2 (January 03, 2024)

* Corrected and added documentation example
* Updated dependancies

## 0.12.1 (October 11, 2023)

* Corrected documentation example
* Provider pushed to terraform registry

## 0.12.0 (October 04, 2023)

ENHANCEMENTS:

* release now with golang 1.21
* resource/**wallix-bastion_user**: update the password when has changed in config to not empty value and `force_change_pwd` isn't true (instead of no-op on password when update resource)

BUG FIXES:

* reduced compute and memory usage to prepare the JSON payload when creating or updating resource

## 0.11.0 (September 26, 2023)

FEATURES:

* add `wallix-bastion_local_password_policy` data source

## 0.10.0 (July 27, 2023)

FEATURES:

* add `wallix-bastion_connection_message` resource

BUG FIXES:

* reduce CRUD operations time (reuse HTTP/TCP connections instead of using a new for each request to API)

## 0.9.1 (May 15, 2023)

BUG FIXES:

* force a resource replacement when `private_key` change on `wallix-bastion_device_localdomain_account_credential` and `wallix-bastion_domain_account_credential` resources (update doesn't work with generated keys)

## 0.9.0 (March 03, 2023)

ENHANCEMENTS:

* resource/**wallix-bastion_profile**: add `dashboards` argument (not compatible with API v3.3)

BUG FIXES:

* fix not detecting that an account's credentials have been deleted while it still exists with resource ID but not linked to the account

## 0.8.0 (February 24, 2023)

FEATURES:

* add `wallix-bastion_configoption` data source

ENHANCEMENTS:

* release now with golang 1.20

## 0.7.0 (January 13, 2023)

FEATURES:

* add `wallix-bastion_authdomain_ad` resource
* add `wallix-bastion_authdomain_azuread` resource
* add `wallix-bastion_authdomain_ldap` resource
* add `wallix-bastion_authdomain_mapping` resource
* add `wallix-bastion_externalauth_saml` resource

ENHANCEMENTS:

* release now with golang 1.19
* optimize resource search when checking if it already exists before create or when importing
* resource/**wallix-bastion_externalauth_ldap**: add `passphrase` argument
* allow use `v3.8` to `api_version` provider argument

BUG FIXES:

* resource/**wallix-bastion_\*domain** & resource/**wallix-bastion_\*credential**: fix missing requirement of `private_key` with `passphrase` argument
* resource/**wallix-bastion_externalauth_kerberos**: fix missing sensitive option on `keytab`
* resource/**wallix-bastion_externalauth_ldap**: fix missing sensitive option on `certificate` and `private_key` and can't be refresh

## 0.6.1 (May 17, 2022)

NOTES:

* use custom User-Agent when request API
* deps: bump terraform-plugin-sdk to v2.16.0

## 0.6.0 (February 25, 2022)

FEATURES:

* add `wallix-bastion_version` data source

ENHANCEMENTS:

* allow use `v3.6` to `api_version` provider argument

BUG FIXES:

* resource/**wallix-bastion_externalauth_kerberos**: add `keytab` argument required in latest version of WAB
* resource/**wallix-bastion_externalauth_radius**: `secret` argument can't be refresh in latest version of WAB
* resource/**wallix-bastion_externalauth_tacacs**: `secret` argument can't be refresh in latest version of WAB

## 0.5.0 (December 9, 2021)

NOTES:

* upgrade golang version to release, so now requires macOS 10.13 High Sierra or later; Older macOS versions are no longer supported.

## 0.4.2 (December 9, 2021)

BUG FIXES:

* resource/**wallix-bastion_connection_policy**: to avoid unnecessary update of resource, `authentication_methods` is now unordered
* resource/**wallix-bastion_application**: avoid large update plan output with unmodified `path` blocks in block set
* resource/**wallix-bastion_targetgroup**: avoid large update plan output with unmodified blocks in block sets

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
