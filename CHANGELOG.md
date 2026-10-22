# changelog

## Unreleased

ENHANCEMENTS:

- **resource/wallix-bastion_domain_account**, **resource/wallix-bastion_device_localdomain_account**, **resource/wallix-bastion_application_localdomain_account**: document that `auto_change_password` and `auto_change_ssh_key` (domain and device accounts only) default to `false` (the Bastion enables them on an account created without them), and that they only matter when the domain has a password change configured

BUG FIXES:

- **resource/wallix-bastion_externalauth_kerberos**: fixed a keytab change keeping the keys of the previous keytab valid: the update merged the new keytab into the existing one, so a changed SPN or a removed encryption type kept the old keys, with no difference in the next plan; the update now replaces the keytab (`force=true`, checked on a Bastion 12.4.3). A keytab merged by an earlier version stays merged until the next update of the resource

## 0.16.0 (October 21, 2026)

BREAKING CHANGES:

- **provider**: these attributes are now sensitive: `current_passphrase` and `new_passphrase` (resource `wallix-bastion_encryption`), `server_private_key` (resource and data source `wallix-bastion_config_x509`), `options` (resource and data source `wallix-bastion_connection_policy`), `options` and `values` (data source `wallix-bastion_configoption`); a configuration that references them in an `output` without `sensitive = true` fails at plan with `Output refers to sensitive values`, and one that builds a `for_each` from them fails with `Invalid for_each argument`; plans show them as `(sensitive value)` (checked when upgrading from 0.15.2 on a Bastion 12.4.3)
- **resource/wallix-bastion_device**: tags are now replaced by those of the configuration, so tags added outside of Terraform are removed by the next apply when the configuration doesn't list them; use `lifecycle { ignore_changes = [tags] }` for tags managed elsewhere
- **resource/wallix-bastion_domain_account_credential**: `propagate_credential_change = true` now requires `password` (or `private_key` for an `ssh_key`, which can't be a `generate:` value), checked at plan
- **resource/wallix-bastion_encryption**: the creation now fails on a Bastion already set up, without sending any request (import the resource instead), the refresh fails while the Bastion is sealed or in an unknown state, and changing only `current_passphrase` no longer sends a passphrase change
- **data-source/wallix-bastion_encryption**: `enabled` now means that the encryption is ready for use (`seal_state` is `unsealed`): a configuration gating on it gets `false` on a sealed Bastion, which it reported as `true` before
- **resource/wallix-bastion_domain**, **resource/wallix-bastion_device_localdomain**: removing `ca_private_key` now fails at plan, `terraform destroy` included; set the key back, or remove the domain from the configuration to destroy it. A key removed with an earlier version is still on the Bastion, and `ca_public_key` shows it

FEATURES:

- add `wallix-bastion_config_x509_crl` resource and data source (CRL revocation check of the X509 authentication certificates, `/config/x509/crl`)
- add `wallix-bastion_config_x509_ocsp` resource and data source (OCSP revocation check of the X509 authentication certificates, `/config/x509/ocsp`)
- add `wallix-bastion_configoption` resource to manage options of a configuration file (requires API v3.12+ and the REST API option `preview_resources` enabled on the Bastion); only the options set in `values` are managed, destroying the resource leaves the options unchanged on the Bastion; numbers in `values` are adapted to the type of the option before each update (a whole number is sent as `60.0` for a `float` option, which the API requires but `jsonencode` never produces, and a number is sent as a string for a `string` or `option` option), and a section set with something other than an object, or an option set with an object, returns a clear error instead of the API internal error

ENHANCEMENTS:

- **resource/wallix-bastion_connection_policy**: accept `WEBAPP` for `protocol` and `type` (API v3.12+); creating a `WEBAPP` policy with an older API version returns an error before the creation request is sent
- **resource/wallix-bastion_connection_policy**: rewrite the documentation, whose examples used authentication methods the provider rejects (`password`, `public_key`, `ntlm`) and options that don't exist; document that `options` must hold the complete set of options returned by the API, otherwise every plan shows a difference, and how to copy an existing policy (e.g. a default one) with the data source while changing some of its options
- **data-source/wallix-bastion_connection_policy**: fix the documentation example (`wallix-bastion_device_service` requires `protocol`) and the listed authentication methods
- **data-source/wallix-bastion_configoption**: add `values` attribute, the option values as a JSON object in the format of the resource `values` argument
- **data-source/wallix-bastion_configoption**: rewrite the documentation, which described arguments and attributes that don't exist (`option_name`, `value`)
- **data-source/wallix-bastion_encryption**: add `seal_state` and `encryption_mode` attributes
- **resource/wallix-bastion_profile**: `target_groups_limitation.default_target_group` is now optional; without it, the Bastion takes the first of `target_groups`

BUG FIXES:

- **provider**: build with Go 1.26.10 (`toolchain` directive in `go.mod`); the release binaries were built with Go 1.26.0, whose standard library has 26 known vulnerabilities reachable from the provider, mostly in `net/http` and `crypto/tls`, plus GO-2026-6809 in `crypto/x509`, fixed in Go 1.26.10 (govulncheck)
- **provider**: reuse the HTTP connections to the Bastion instead of opening a new one for every API request and leaving it open (one client per provider instance, at most 10 connections, idle ones closed after 4 seconds); with a few dozen API calls within seconds, a Bastion 12.4 stopped accepting new connections and requests failed at random with `context deadline exceeded`
- **resource/wallix-bastion_application_localdomain_account**: fixed every update wiping the password of the account; the update replaces the credentials of the account (`force=true`) and always sent them, empty when `password` isn't set (the documented setup with a `wallix-bastion_application_localdomain_account_credential` resource); they are now only sent on creation and when `password` changes
- **resource/wallix-bastion_user**, **resource/wallix-bastion_timeframe** (and their data sources): escape the name in the request path; a name with `#` or `?` was cut there, so that destroying the user `a#b` deleted the user `a` (checked on a Bastion 12.4), and updating it updated `a`
- **resource/wallix-bastion_application**: fixed emptying `parameters` being ignored, with the same change planned after every apply; the emptied value is now sent on update, which clears it (checked on a Bastion 12.4.3)
- **resource/wallix-bastion_user**: fixed changing `preferred_language` or `force_change_pwd` on an existing user, which was only sent at creation: the language change was planned again after every apply, and the `force_change_pwd` change was saved in the state without reaching the Bastion; both are now sent on update when they change
- **resource/wallix-bastion_application**: fixed the destroy failing with `409 The given application is still associated with one target` when an account of one of its global domains was attached to it (for example by a target group session account; destroying the target group leaves the attachment): on this conflict, the application is detached from the accounts of its global domains, then deleted
- **resource/wallix-bastion_profile**: fixed the import setting `profile_name` to the API id, which forced the replacement of the profile on the next plan; `profile_name` is now read from the API
- **resource/wallix-bastion_authorization**: fixed `authorize_session_sharing = false` never being applied, which left session sharing enabled on the Bastion; disabling it is now sent on update (without the mode, and enabling or untouched sharing send the same requests as before); while sharing is disabled, the mode the Bastion keeps is no longer reported as a change
- **resource/wallix-bastion_domain_account**: fixed `resources = []` never being applied, which left the account attached to its resources on the Bastion; an emptied list is now sent on update (when `resources` isn't set, the resources in state are still sent unchanged)
- **resource/wallix-bastion_device**: fixed removing or changing a tag: without `force=true`, the API adds the tags sent to the existing ones, so a removed tag stayed on the device and a changed value failed with `409 cannot have multiple Tags with the same keys`; the update now replaces the tags with those of the configuration (checked on a Bastion 12.4.1)
- **resource/wallix-bastion_profile**, **resource/wallix-bastion_application**: fixed `dashboards = []` (profile) and removing the last tag (application) being ignored, with the same change planned after every apply; the emptied list is now sent
- **resource/wallix-bastion_device_service**: the emptied `subprotocols` of an RDP service is now sent, which clears them
- **resource/wallix-bastion_authorization**: removing the last `subprotocols` was ignored, with the same change planned after every apply; the Bastion can't clear them, even with `authorize_sessions = false` (`400 At least one subprotocol is required`), so it now replaces the authorization
- **resource/wallix-bastion_device_service**: an SSH service without `subprotocols` is now refused at plan; the Bastion refuses it at creation and on update (checked on a Bastion 12.4.1)
- **resource/wallix-bastion_timeframe**: fixed renaming a timeframe failing with `Root object was present, but now absent` and leaving it on the Bastion out of state; the API renames it, and the resource id (the name) now follows the new name
- **resource/wallix-bastion_domain**, **resource/wallix-bastion_device_localdomain**: fixed changing `ca_private_key` to a `generate:` value being ignored when the domain already had a key (a previous `generate:` value or a PEM key); the Bastion kept its key while the state recorded the new value. An unchanged `generate:` value is still not sent, since the Bastion generates a new key on every request that holds one
- **resource/wallix-bastion_domain**, **resource/wallix-bastion_device_localdomain**: `ca_public_key` is now known after apply when `ca_private_key` changes; it was planned unchanged, so the resources using it got the previous key in the apply that replaced it, and needed a second apply
- **resource/wallix-bastion_domain**, **resource/wallix-bastion_device_localdomain**: removing `ca_private_key` is now refused at plan; the Bastion can't remove the CA key (an empty value is ignored, `null` is refused), so the key stayed active while the state recorded its removal
- **resource/wallix-bastion_connection_policy**: changing `protocol` now replaces the policy, like `type`; `protocol` is only sent on creation, so the change was never applied and the plan showed it again after every apply
- **resource/wallix-bastion_encryption**: mark `current_passphrase` and `new_passphrase` as sensitive; they were shown in plans
- **resource/wallix-bastion_config_x509**, **data-source/wallix-bastion_config_x509**: mark `server_private_key` as sensitive; the private key was shown in plans
- **resource/wallix-bastion_configoption**, **data-source/wallix-bastion_configoption**: mark `values` (and `options` of the data source) as sensitive; the API returns the configuration options in clear, keys included (`wabengine` `backup_key`, `wabrestapi` `data_encryption_key`, `sessionlogpolicy` `archive_key`, checked on a Bastion 12.4)
- **resource/wallix-bastion_connection_policy**, **data-source/wallix-bastion_connection_policy**: mark `options` as sensitive; they hold passwords (`proxy.password`, `vnc_over_ssh.ssh_password`, `rdp.krb_armoring_fallback_password`), which the API returns in clear (checked on a Bastion 12.4.3)
- **resource/wallix-bastion_domain_account_credential**: after a rotation recreated the credential with a new id, the read dropped it from state and every following apply failed with `already exists`; the read now looks the credential up by account and type, as the device and application variants do since 0.15.0, and a deleted account no longer makes the read fail
- **resource/wallix-bastion_device_localdomain_account_credential**, **resource/wallix-bastion_application_localdomain_account_credential**: fixed every plan failing with `api doesn't return OK: 404` once the account, local domain, device or application of the credential was deleted outside of Terraform; the credential is now dropped from state, as the domain variant does
- **provider**: escape the name in the API searches by name (`?q=<field>=<name>`) that the resources run before a creation and on import, and that the data sources use; a name with `&`, `+`, `#` or a space was cut or altered (`a+b` was searched as `a b` on a Bastion 12.4), so the creation of `a&b` failed with `already exists` when `a` existed, and its import imported `a`; the name is also escaped for the Bastion search expression, where a name with `&&` or `||` failed with `Invalid expression` and one with `\` was not found
- **provider**: a destroy no longer fails with `404` on an object already deleted outside of Terraform, which happened when the state wasn't refreshed first (`-refresh=false`, a saved plan applied later, or a child deleted by the Bastion with its parent) and left the objects depending on it; every resource now treats a `404` on delete as done, except a `404` for an unknown API version or route, which stays an error as the object may still exist, and the credential resources first delete a credential rotated under a new id since the last refresh, so that it doesn't stay on the Bastion
- **resource/wallix-bastion_domain_account_credential**: fixed `propagate_credential_change = true` sending an empty body, or the password route for an SSH key, to `/accountchangepassword`: the Bastion then generated a new random credential and set it on the target host (checked on a Bastion 12.4.1); the `password`, or the `private_key` and `passphrase`, of the configuration is now sent on the route of the credential type, only when it changes, and a credential whose propagation fails stays in state (tainted) instead of being left on the Bastion out of state
- **resource/wallix-bastion_encryption**, **data-source/wallix-bastion_encryption**: fixed the encryption state never being read (a switch on API versions without their `v` prefix, and fields that don't exist): `enabled` was always `true`, even on a sealed Bastion; `seal_state` is now read, and setting up without a passphrase (`new_passphrase = ""`) now sends it instead of an empty body
- **resource/wallix-bastion_profile**, **data-source/wallix-bastion_profile**: fixed the provider crashing when reading a profile whose target groups limitation has no default target group (`null` in the API), for example an enabled limitation without target groups
- **resource/wallix-bastion_externalauth_saml**: fixed the resource being unusable on API v3.8 since 0.15.0: `claim_customization` was required and always sent, and API v3.8 refuses it (`Additional properties are not allowed ('claim_customization' was unexpected)`, checked on a Bastion 12.4.1); it is now required at plan on API v3.12 and later only, refused at plan on v3.8, and never sent to v3.8

NOTES:

- **resource/wallix-bastion_connection_policy**: if your plan showed the same `protocol` change after every apply, the next plan replaces the policy (destroy, then create, under the same name); the destroy fails while devices or applications still use the policy
- **resource/wallix-bastion_user**: `force_change_pwd` isn't read from the Bastion, so a change of `force_change_pwd` applied with an earlier version is in the state but not on the Bastion, and the next plan doesn't show it; to apply it, set the opposite value and apply, then set it back and apply again
- **resource/wallix-bastion_profile**: `profile_name` is now read on every refresh, so a profile renamed outside Terraform shows as a change, which replaces the profile
- **resource/wallix-bastion_domain**, **resource/wallix-bastion_device_localdomain**: changing `ca_private_key` to a `generate:` value now replaces the CA key of the domain, so `ca_public_key` changes and the targets that trust the previous key must be updated; a change made before this fix shows no difference while the Bastion still has the previous key, so change the value again to get a new key
- **resource/wallix-bastion_application_localdomain_account**: removing `password` from the configuration no longer clears the password on the Bastion

## 0.15.2 (October 01, 2026)

ENHANCEMENTS:

- **resource/wallix-bastion_application**: add `login_form_url`, `login_button_selector`, and `allow_non_post_form` arguments (API v3.12+, `category = "web_application"` only) to configure the automatic injection of credentials into the application's login form; they are only sent to the API when set, or on update to clear them, so existing configurations send the same requests as before
- **data-source/wallix-bastion_application**: add `login_form_url`, `login_button_selector`, and `allow_non_post_form` attributes

BUG FIXES:

- **resource/wallix-bastion_application**: fixed updating an application with `category = "web_application"`, which always failed with `application_url cannot be configured when category = standard`: the category was only read at creation, so the update was validated and built as a standard application; the update now follows the resource category, and `category` is still only sent at creation since it forces a new resource

NOTES:

- **resource/wallix-bastion_application**: as a side effect, the `jumphost` branch is now reachable on create and update. On API versions below v3.12, a create sends `application_url` and `browser` without `category` (which is only sent from v3.12) instead of being rejected client-side; on API v3.12+, updating a legacy `jumphost` application now returns the existing deprecation error. Both paths were already unsupported

## 0.15.1 (September 16, 2026)

BUG FIXES:

- **resource/wallix-bastion_application**: fixed `parameters` always being sent to the API, even when left unset, and unconditionally included when `category = "web_application"`, where the API rejects the key outright and rejects the whole request; `parameters` is now omitted from the request unless explicitly set, and setting it with `category = "web_application"` now returns a clear error instead of a rejected API call

## 0.15.0 (August 24, 2026)

BREAKING CHANGES:

- **provider**: TLS certificate verification is now enabled by default (`insecure_skip_verify` now defaults to `false`). Users connecting to a Bastion with a self-signed certificate must explicitly set `insecure_skip_verify = true` or the `WALLIX_INSECURE_SKIP_VERIFY` environment variable.
- **resource/wallix-bastion_externalauth_saml**: `claim_customization` is now required, matching the API's actual requirement on API v3.12+; existing configs that omit it must add the block. This resource is not usable on API v3.8, where `claim_customization` isn't part of the API at all.

FEATURES:

- add `wallix-bastion_apikey` resource
- add `wallix-bastion_apikey_v2` resource (requires API v3.12+; adds a required `profile` argument on top of `wallix-bastion_apikey`)
- add `wallix-bastion_application_localdomain_account_credential` resource (only `type = "password"` is supported by the API for this endpoint)
- add `wallix-bastion_certificate_authority` resource (requires API v3.12+)
- add `wallix-bastion_passwordchangepolicy` resource
- add `wallix-bastion_notification` resource
- add `wallix-bastion_config_smtp` resource (SMTP server configuration)
- add `wallix-bastion_config_wsm` resource (Web Session Manager configuration, requires API v3.12+)
- **provider**: add `session_timeout`, `csrf_enabled`, and `insecure_skip_verify` arguments
- **client**: cookie-based session authentication, avoiding re-authentication on every API request
- **client**: CSRF token protection, with automatic extraction from responses and refresh on expiry
- add 39 data sources for read-only lookup of infrastructure that already exists on the Bastion,
  covering every resource that previously had none (`wallix-bastion_device`,
  `wallix-bastion_application`, `wallix-bastion_user`, `wallix-bastion_usergroup`,
  `wallix-bastion_targetgroup`, `wallix-bastion_cluster`, `wallix-bastion_profile`,
  `wallix-bastion_authorization`, and 31 others including the full `device_*`/`domain_*`/
  `application_*` nested-resource families, `externalauth_*`, `authdomain_*`, and the singleton
  configs `wallix-bastion_config_x509`/`wallix-bastion_config_smtp`/`wallix-bastion_config_wsm`/
  `wallix-bastion_encryption`)
- default API now use v3.12

ENHANCEMENTS:

- reduced resource creation time across almost every resource: creation now reads the new resource's ID from the API's `X-Object-Id` response header instead of an extra search request afterward, falling back to the search only against older Bastion versions that don't return the header
- **resource/wallix-bastion_device**, **resource/wallix-bastion_application**: add `tags` argument (repeatable `key`/`value` blocks)
- **resource/wallix-bastion_application**: add `web_application` category, replacing the `jumphost` category deprecated on API v3.12+
- dependency updates addressing multiple security advisories in `golang.org/x/net`, `golang.org/x/crypto`, `google.golang.org/grpc`, and `github.com/cloudflare/circl`
- **ci**: migrated `golangci-lint` configuration to the v2 schema and updated CI Go versions to 1.25/1.26
- **ci**: migrated acceptance tests off the deprecated `resource.TestCase.Providers` field to `ProviderFactories`, and resource importers off the deprecated `schema.ResourceImporter.State` to `StateContext`, clearing 129 `staticcheck` SA1019 findings
- **ci**: introduced shared string constants for schema key names and other repeated literals, replacing ~1300 duplicate occurrences across resources/data sources, and re-enabled the `goconst` linter

BUG FIXES:

- **resource/wallix-bastion_domain_account**: fixed a bug where changing the domain account without also updating its associated resources caused those resources to be deleted
- **resource/wallix-bastion_application**: fixed a regression in the web application category where the default category wasn't applied when left unset
- **resource/wallix-bastion_device_localdomain_account**: fixed a potential panic when the API omits the `credentials` field from a read response
- **resource/wallix-bastion_authdomain_azuread**: fixed `passphrase` always being sent to the API as an empty string when unset, which made it impossible to create this resource without also setting `private_key`/`passphrase`, even though both are documented as optional
- **resource/wallix-bastion_device_localdomain_account_credential**, **resource/wallix-bastion_application_localdomain_account_credential**: fixed `apply` failing with "already exists" after WALLIX rotates a credential outside of Terraform (manual regenerate, or an automatic rotation policy); `Read` now retries a lookup by `(account, type)` before treating a 404 on the stored ID as a deletion
- **data-source/wallix-bastion_device_localdomain_account_credential**, **data-source/wallix-bastion_domain_account_credential**, **data-source/wallix-bastion_application_localdomain_account_credential**: fixed `password`/`passphrase`/`private_key` never being populated even though declared as computed attributes
- **client**: fixed a race condition on session state checks, a bug where POST/PUT requests could fail on re-authentication due to request body reuse, and defensive error handling around URL parsing and cookie jar creation
- **docs**: rewrote the `wallix-bastion_authdomain_ad`, `wallix-bastion_authdomain_azuread`, `wallix-bastion_authdomain_ldap`, `wallix-bastion_authdomain_mapping`, `wallix-bastion_authdomain_saml`, `wallix-bastion_externalauth_kerberos`, `wallix-bastion_externalauth_ldap`, `wallix-bastion_externalauth_radius`, `wallix-bastion_externalauth_saml`, and `wallix-bastion_externalauth_tacacs` resource docs, whose examples referenced arguments that don't exist on these resources (e.g. `identity_provider_url`, `user_attribute_mapping`, `ldap_hosts`, `tenant_id`) instead of their real schemas

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
