# changelog

## 0.15.0 (July 23, 2026)

BREAKING CHANGES:

- **provider**: TLS certificate verification is now enabled by default (`insecure_skip_verify` now defaults to `false`). Users connecting to a Bastion with a self-signed certificate must explicitly set `insecure_skip_verify = true` or the `WALLIX_INSECURE_SKIP_VERIFY` environment variable.

FEATURES:

- add `wallix-bastion_apikey` resource
- add `wallix-bastion_apikey_v2` resource (requires API v3.12+; adds a required `profile` argument on top of `wallix-bastion_apikey`)
- add `wallix-bastion_application_localdomain_account_credential` resource (only `type = "password"` is supported by the API for this endpoint)
- add `wallix-bastion_certificate_authority` resource (requires API v3.12+)
- add `wallix-bastion_passwordchangepolicy` resource
- add `wallix-bastion_notification` resource
- **provider**: add `session_timeout`, `csrf_enabled`, and `insecure_skip_verify` arguments
- **client**: cookie-based session authentication, avoiding re-authentication on every API request
- **client**: CSRF token protection, with automatic extraction from responses and refresh on expiry
- add 37 data sources for read-only lookup of infrastructure that already exists on the Bastion,
  covering every resource that previously had none (`wallix-bastion_device`,
  `wallix-bastion_application`, `wallix-bastion_user`, `wallix-bastion_usergroup`,
  `wallix-bastion_targetgroup`, `wallix-bastion_cluster`, `wallix-bastion_profile`,
  `wallix-bastion_authorization`, and 29 others including the full `device_*`/`domain_*`/
  `application_*` nested-resource families, `externalauth_*`, `authdomain_*`, and both
  singleton configs `wallix-bastion_config_x509`/`wallix-bastion_encryption`)

ENHANCEMENTS:

- reduced resource creation time across almost every resource: creation now reads the new resource's ID from the API's `X-Object-Id` response header instead of an extra search request afterward, falling back to the search only against older Bastion versions that don't return the header
- **resource/wallix-bastion_device**, **resource/wallix-bastion_application**: add `tags` argument (repeatable `key`/`value` blocks)
- **resource/wallix-bastion_application**: add `web_application` category, replacing the `jumphost` category deprecated on API v3.12+
- dependency updates addressing multiple security advisories in `golang.org/x/net`, `golang.org/x/crypto`, `google.golang.org/grpc`, and `github.com/cloudflare/circl`
- **ci**: migrated `golangci-lint` configuration to the v2 schema and updated CI Go versions to 1.25/1.26

BUG FIXES:

- **resource/wallix-bastion_domain_account**: fixed a bug where changing the domain account without also updating its associated resources caused those resources to be deleted
- **resource/wallix-bastion_application**: fixed a regression in the web application category where the default category wasn't applied when left unset
- **resource/wallix-bastion_device_localdomain_account**: fixed a potential panic when the API omits the `credentials` field from a read response
- **resource/wallix-bastion_authdomain_azuread**: fixed `passphrase` always being sent to the API as an empty string when unset, which made it impossible to create this resource without also setting `private_key`/`passphrase`, even though both are documented as optional
- **client**: fixed a race condition on session state checks, a bug where POST/PUT requests could fail on re-authentication due to request body reuse, and defensive error handling around URL parsing and cookie jar creation

## 0.14.8 (October 10, 2025)

BUG FIXES:

- **provider**: improved error handling and validation when configuring provider with missing or invalid values (ip, user, port) to provide clearer error messages instead of panics.
- **provider**: fixed potential panic when making HTTP requests with bad client configuration by improving error handling order.
- **resource/wallix-bastion_config_x509**: fixed infinite non-empty plan issue by avoiding to set the full API response in state and instead comparing certificate common names to detect changes.
- **resource/wallix-bastion_config_x509**: fixed update of `enable` attribute from `true` to `false` and ensure it's properly set in state only when configured.
- **resource/wallix-bastion_config_x509**: added 3-second delay after modifying X509 configuration to wait for API listener restart with new certificate.

ENHANCEMENTS:

- **resource/wallix-bastion_device_service**: added a format control on service_name to avoid API error when creating/updating a service with invalid characters.
- **ci**: updated GitHub workflows with improved linting configuration and dependency updates.
- **ci**: enhanced golangci-lint configuration with proper import ordering and deprecation handling.
- **docs**: updated documentation workflow with better validation and formatting checks.

## 0.14.7 (September 25, 2025)

BUG FIXES:

- **ci_release**: fixed an issue where the release process did not correctly build artefact and used tar.gz instead of zip.

## 0.14.6 (June 14, 2025)

BUG FIXES:

- **resource/wallix-bastion_authorization**: fixed an issue where session sharing arguments were not correctly validated when omitted.
- **resource/wallix-bastion_config_x509**: improved error handling for invalid certificate chains.

ENHANCEMENTS:

- **provider**: improved logging for API request failures to aid debugging.
- **docs**: updated documentation for session sharing options and credenal propagation.

## 0.14.5 (August 25, 2025)

ENHANCEMENTS:

- **resource/wallix-bastion_authorization**: added support for session sharing functionality with new
  `authorize_session_sharing` (boolean) and `session_sharing_mode` (enum: "view_only", "view_control")
  arguments, enabling users to configure session sharing permissions for authorizations.

## 0.14.4 (March 3, 2025)

FEATURES:

- **resource/wallix-bastion_authdomain_saml**: added the possibilty to configure Other IDPs/SAML auth domain resource.

BUG FIXES:

## 0.14.3 (March 3, 2025)

BUG FIXES:

- **resource/wallix-config_x509**: Fixed previous build not including the resource

## 0.14.2 (December 20, 2024)

FEATURES:

- **resource/wallix-bastion_config_x509**: added the possibilty to configure the X509 for the GUI and for users authentication

BUG FIXES:

- **resource/wallix-device_service**: supported subprotocols.

## 0.14.1 (December 13, 2024)

FEATURES:

- **datasource/wallix-bastion_authdomain_ad**: added the datasource to retrieve an existing authdomain
- **resource/wallix-bastion_domain_account_credential**: added credential propagation to AD upon creation.

BUG FIXES:

- **resource/wallix-bastion_externalauth_kerberos**: deprecate `login_attribute` argument (it produces Bad Request with API v3.12)
- **provider_test**: Added the user environment variable presence test for acceptance tests.

## 0.14.0 (November 08, 2024)

BREAKING CHANGES:

- remove compatibility with API version 3.3 and 3.6
- remove resource `wallix-bastion_ldapdomain`
- remove resource `wallix-bastion_ldapmapping`
- default provider api_version argument is now `v3.8`
- user statement is now mandatory

FEATURES:

- add compatibility with API version 3.12

ENHANCEMENTS:

- **resource/wallix-bastion_application**:
  - add `category`, `application_url`, `browser`, `browser_version` arguments to be able to add `jumphost` application (not tested)
  - `paths` and `target` is now only required when `category` = `standard`
- **resource/wallix-bastion_connection_policy**: add `type` argument with default value as `protocol` value
- **resource/wallix-bastion_externalauth_saml**: add `claim_customization` block argument

## 0.13.0 (March 08, 2024)

- build(deps): bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 by @dependabot in https://github.com/wallix/terraform-provider-wallix-bastion/pull/13
- added http basic authentication by @moulip in https://github.com/wallix/terraform-provider-wallix-bastion/pull/15

## 0.12.2 (January 03, 2024)

- Corrected and added documentation example
- Updated dependancies

## 0.12.1 (October 11, 2023)

- Corrected documentation example
- Provider pushed to terraform registry

## 0.12.0 (October 04, 2023)

ENHANCEMENTS:

- release now with golang 1.21
- resource/**wallix-bastion_user**: update the password when has changed in config to not empty value and `force_change_pwd` isn't true (instead of no-op on password when update resource)

BUG FIXES:

- reduced compute and memory usage to prepare the JSON payload when creating or updating resource

## 0.11.0 (September 26, 2023)

FEATURES:

- add `wallix-bastion_local_password_policy` data source

## 0.10.0 (July 27, 2023)

FEATURES:

- add `wallix-bastion_connection_message` resource

BUG FIXES:

- reduce CRUD operations time (reuse HTTP/TCP connections instead of using a new for each request to API)

## 0.9.1 (May 15, 2023)

BUG FIXES:

- force a resource replacement when `private_key` change on `wallix-bastion_device_localdomain_account_credential` and `wallix-bastion_domain_account_credential` resources (update doesn't work with generated keys)

## 0.9.0 (March 03, 2023)

ENHANCEMENTS:

- resource/**wallix-bastion_profile**: add `dashboards` argument (not compatible with API v3.3)

BUG FIXES:

- fix not detecting that an account's credentials have been deleted while it still exists with resource ID but not linked to the account

## 0.8.0 (February 24, 2023)

FEATURES:

- add `wallix-bastion_configoption` data source

ENHANCEMENTS:

- release now with golang 1.20

## 0.7.0 (January 13, 2023)

FEATURES:

- add `wallix-bastion_authdomain_ad` resource
- add `wallix-bastion_authdomain_azuread` resource
- add `wallix-bastion_authdomain_ldap` resource
- add `wallix-bastion_authdomain_mapping` resource
- add `wallix-bastion_externalauth_saml` resource

ENHANCEMENTS:

- release now with golang 1.19
- optimize resource search when checking if it already exists before create or when importing
- resource/**wallix-bastion_externalauth_ldap**: add `passphrase` argument
- allow use `v3.8` to `api_version` provider argument

BUG FIXES:

- resource/**wallix-bastion\_\*domain** & resource/**wallix-bastion\_\*credential**: fix missing requirement of `private_key` with `passphrase` argument
- resource/**wallix-bastion_externalauth_kerberos**: fix missing sensitive option on `keytab`
- resource/**wallix-bastion_externalauth_ldap**: fix missing sensitive option on `certificate` and `private_key` and can't be refresh

## 0.6.1 (May 17, 2022)

NOTES:

- use custom User-Agent when request API
- deps: bump terraform-plugin-sdk to v2.16.0

## 0.6.0 (February 25, 2022)

FEATURES:

- add `wallix-bastion_version` data source

ENHANCEMENTS:

- allow use `v3.6` to `api_version` provider argument

BUG FIXES:

- resource/**wallix-bastion_externalauth_kerberos**: add `keytab` argument required in latest version of WAB
- resource/**wallix-bastion_externalauth_radius**: `secret` argument can't be refresh in latest version of WAB
- resource/**wallix-bastion_externalauth_tacacs**: `secret` argument can't be refresh in latest version of WAB

## 0.5.0 (December 9, 2021)

NOTES:

- upgrade golang version to release, so now requires macOS 10.13 High Sierra or later; Older macOS versions are no longer supported.

## 0.4.2 (December 9, 2021)

BUG FIXES:

- resource/**wallix-bastion_connection_policy**: to avoid unnecessary update of resource, `authentication_methods` is now unordered
- resource/**wallix-bastion_application**: avoid large update plan output with unmodified `path` blocks in block set
- resource/**wallix-bastion_targetgroup**: avoid large update plan output with unmodified blocks in block sets

## 0.4.1 (October 18, 2021)

ENHANCEMENTS:

- [docs] reformat arguments/attributes, add example usage & minor fix

BUG FIXES:

- fix the potential double slash in url when calling Wallix API
- fix missing sensitive options for few arguments
- resource/**wallix-bastion_application_localdomain**, **wallix-bastion_device_localdomain**, **wallix-bastion_domain**: fix arguments requirement
- resource/**wallix-bastion_application**: fix panic with `global_domains`
- resource/**wallix-bastion_profile**: fix `default_target_group` is required in `target_groups_limitation` block
- resource/**wallix-bastion_domain**: fix `passphrase` can't refresh
- resource/**wallix-bastion_device_localdomain**: fix `passphrase` can't refresh

## 0.4.0 (April 9, 2021)

FEATURES:

- add `wallix-bastion_domain` data source

## 0.3.3 (April 6, 2021)

BUG FIXES:

- fix `global_domains` argument can be an attribute in `wallix-bastion_device_service` resource

## 0.3.2 (April 1, 2021)

BUG FIXES:

- fix `device`/`service` or `application` needed with `domain_type`="global" on `session_accounts` in `wallix-bastion_targetgroup` resource
- fix `resources` argument can be an attribute in `wallix-bastion_domain_account` resource

## 0.3.1 (March 30, 2021)

BUG FIXES:

- fix import user resource

## 0.3.0 (March 19, 2021)

FEATURES:

- add `wallix-bastion_application` resource
- add `wallix-bastion_application_localdomain` resource
- add `wallix-bastion_application_localdomain_account` resource
- add `wallix-bastion_checkout_policy` resource
- add `wallix-bastion_cluster` resource
- add `wallix-bastion_connection_policy` resource
- add `wallix-bastion_externalauth_kerberos` resource
- add `wallix-bastion_externalauth_radius` resource
- add `wallix-bastion_externalauth_tacacs` resource
- add `wallix-bastion_profile` resource
- add `wallix-bastion_timeframe` resource

## 0.2.0 (March 5, 2021)

FEATURES:

- add `wallix-bastion_authorization` resource
- add `wallix-bastion_device`resource
- add `wallix-bastion_device_localdomain` resource
- add `wallix-bastion_device_localdomain_account` resource
- add `wallix-bastion_device_localdomain_account_credential` resource
- add `wallix-bastion_device_service` resource
- add `wallix-bastion_domain` resource
- add `wallix-bastion_domain_account` resource
- add `wallix-bastion_domain_account_credential` resource
- add `wallix-bastion_ldapdomain` resource
- add `wallix-bastion_ldapmapping` resource
- add `wallix-bastion_targetgroup` resource

ENHANCEMENTS:

- remove Forcenew on `authentication_name` in `wallix-bastion_externalauth_ldap` resource, it's not necessary

BUG FIXES:

- typo in errors displayed
- remove log to debug in http request (possible secret could appear)
- `timeframes` and `restrictions` aren't ordered in `wallix-bastion_usegroup` resource
- do not reactivate `force_change_pwd` after creation and the password has changed in `wallix-bastion_user` resource

## 0.1.0 (February 9, 2021)

First release
