<!-- markdownlint-disable-file MD013 MD041 -->
FEATURES:

* **client**: Implement cookie-based session authentication to avoid re-authenticating on every API request

ENHANCEMENTS:

* **client**: Add helper method `getHTTPClient()` to optimize HTTP client creation and avoid inefficient copying
* **client**: Add helper method `buildURL()` to centralize URL construction and reduce code duplication
* **client**: Extract duplicated string literals into constants for better maintainability
* **client**: Add automatic re-authentication on 401 responses with proper session state management
* **client**: Improve error messages for authentication failures

BUG FIXES:

* **client**: Fix race condition on session state check by adding mutex protection
* **client**: Fix critical bug where POST/PUT requests would fail on re-authentication due to io.Reader reuse
* **client**: Add proper error handling in retry logic to prevent potential panics
* **client**: Validate HTTP status code before extracting session cookies
* **client**: Handle URL parsing and cookie jar creation errors defensively
