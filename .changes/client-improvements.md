<!-- markdownlint-disable-file MD013 MD041 -->
FEATURES:

* **client**: Implement cookie-based session authentication to avoid re-authenticating on every API request
* **client**: Implement CSRF token protection with automatic extraction and application to all requests
* **provider**: Add `insecure_skip_verify` option for TLS certificate verification control (default: secure)
* **provider**: Add `csrf_enabled` option to enable/disable CSRF token protection (default: enabled)
* **provider**: Add `session_timeout` option to configure session timeout duration

ENHANCEMENTS:

* **client**: Add helper method `getHTTPClient()` to optimize HTTP client creation and avoid inefficient copying
* **client**: Add helper method `buildURL()` to centralize URL construction and reduce code duplication
* **client**: Extract duplicated string literals into constants for better maintainability
* **client**: Add automatic re-authentication on 401 responses with proper session state management
* **client**: Add automatic CSRF token refresh on 403 responses with retry logic
* **client**: Improve error messages for authentication failures
* **client**: Support CSRF token extraction from both `X-CSRF-Token` header and `api_csrf_token` cookie
* **client**: Implement configurable TLS settings with secure-by-default behavior
* **provider**: TLS certificate verification is now secure by default (InsecureSkipVerify=false)
* **tests**: Add comprehensive CSRF integration tests covering token extraction, refresh, and retry logic
* **tests**: Add TLS configuration tests validating secure defaults and environment variable support
* **tests**: Make X.509 certificate tests dynamic using configured Bastion hostname

BUG FIXES:

* **client**: Fix race condition on session state check by adding mutex protection
* **client**: Fix critical bug where POST/PUT requests would fail on re-authentication due to io.Reader reuse
* **client**: Add proper error handling in retry logic to prevent potential panics
* **client**: Validate HTTP status code before extracting session cookies
* **client**: Handle URL parsing and cookie jar creation errors defensively
* **client**: Remove obsolete `defaultHTTPClient` global variable that bypassed TLS configuration
* **data_source_version**: Fix TLS bypass by using `getHTTPClient()` instead of hardcoded client
* **data_source_version**: Refactor to use `c.newRequest()` for consistency with other data sources and resources, enabling session management, CSRF protection, and automatic retry logic

BREAKING CHANGES:

* **provider**: TLS certificate verification is now enabled by default. Users working with self-signed certificates must explicitly set `insecure_skip_verify = true` or use the `WALLIX_INSECURE_SKIP_VERIFY` environment variable
