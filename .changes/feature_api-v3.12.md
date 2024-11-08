<!-- markdownlint-disable-file MD013 MD041 -->
BREAKING CHANGES:

* remove compatibility with API version 3.3 and 3.6
* remove resource `wallix-bastion_ldapdomain`
* remove resource `wallix-bastion_ldapmapping`
* default provider api_version argument is now `v3.8`

FEATURES:

* add compatibility with API version 3.12

ENHANCEMENTS:

* **resource/wallix-bastion_application**:
  * add `category`, `application_url`, `browser`, `browser_version` arguments to be able to add `jumphost` application (not tested)
  * `paths` and `target` is now only required when `category` = `standard`
* **resource/wallix-bastion_connection_policy**: add `type` argument with default value as `protocol` value
* **resource/wallix-bastion_externalauth_saml**: add `claim_customization` block argument
