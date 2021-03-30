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
