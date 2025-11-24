package bastion

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

const (
	userAgentHeader   = "User-Agent"
	userAgentValue    = "terraform-provider-wallix-bastion"
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	csrfTokenHeader   = "X-CSRF-Token"
	csrfCookieName    = "api_csrf_token"
)

type Client struct {
	bastionPort        int
	bastionAPIVersion  string
	bastionIP          string
	bastionToken       string
	bastionUser        string
	bastionPwd         string
	insecureSkipVerify bool

	jar             http.CookieJar
	cookie          http.Cookie
	isAuthenticated bool
	csrfToken       string
	sessionTimeout  int
	csrfEnabled     bool
	authMu          sync.Mutex
}

func (c *Client) initJar() {
	if c.jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			// This should never happen with nil options, but panic if it does
			panic(fmt.Sprintf("failed to create cookie jar: %v", err))
		}
		c.jar = jar
	}
}

func (c *Client) getHTTPClient() *http.Client {
	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.insecureSkipVerify}

	return &http.Client{
		Transport: transport,
		Jar:       c.jar,
	}
}

func (c *Client) buildURL(path string) string {
	host := net.JoinHostPort(c.bastionIP, strconv.Itoa(c.bastionPort))

	return fmt.Sprintf("https://%s%s", host, path)
}

func (c *Client) cookieValid() bool {
	if c.cookie.Name != "wab_session_id" || c.cookie.Value == "" {
		return false
	}

	return true
}

func (c *Client) authenticate(ctx context.Context) error {
	c.authMu.Lock()

	defer c.authMu.Unlock()

	if c.isAuthenticated && c.cookieValid() {
		return nil
	}

	c.initJar()

	authURL := c.buildURL("/api")
	client := c.getHTTPClient()
	authCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(authCtx, http.MethodPost, authURL, nil)
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}

	req.Header.Set(userAgentHeader, userAgentValue)
	req.Header.Set("Accept", contentTypeJSON)
	req.Header.Set(contentTypeHeader, contentTypeJSON)

	if c.bastionToken != "" {
		req.Header.Set("X-Auth-Key", c.bastionToken)
		req.Header.Set("X-Auth-User", c.bastionUser)
	} else {
		req.SetBasicAuth(c.bastionUser, c.bastionPwd)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check for non-2xx status code before processing cookies
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("authentication failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	u, err := url.Parse(authURL)
	if err != nil {
		return fmt.Errorf("parsing authURL: %w", err)
	}
	cookies := c.jar.Cookies(u)

	var found *http.Cookie
	for _, ck := range cookies {
		if ck.Name == "wab_session_id" {
			tmp := *ck
			found = &tmp

			break
		}
	}

	if found == nil {
		return errors.New("authentication failed: no wab_session_id cookie returned")
	}

	c.cookie = *found

	// Extract CSRF token from response headers or cookies
	c.extractCSRFToken(resp)

	c.isAuthenticated = true

	return nil
}

func (c *Client) extractCSRFToken(resp *http.Response) {
	if !c.csrfEnabled {
		return
	}

	// Try to extract from X-CSRF-Token header first.
	if token := resp.Header.Get(csrfTokenHeader); token != "" {
		c.csrfToken = token

		return
	}

	// Fall back to api_csrf_token cookie (Bastion 12.0.3+).
	for _, cookie := range resp.Cookies() {
		if cookie.Name == csrfCookieName {
			c.csrfToken = cookie.Value

			return
		}
	}

	// CSRF token is optional - some Bastion versions don't require it.
}

func (c *Client) addCSRFHeader(req *http.Request) {
	if c.csrfEnabled && c.csrfToken != "" {
		req.Header.Set(csrfTokenHeader, c.csrfToken)
	}
}

func (c *Client) newRequest(ctx context.Context, uri, method string, jsonBody interface{}) (string, int, error) {
	c.authMu.Lock()
	needAuth := !c.isAuthenticated || !c.cookieValid()
	c.authMu.Unlock()

	if needAuth {
		if err := c.authenticate(ctx); err != nil {
			return "", http.StatusUnauthorized, fmt.Errorf("authentication failed: %w", err)
		}
	}

	c.initJar()

	urlStr := c.buildURL("/api/" + c.bastionAPIVersion)
	if strings.HasPrefix(uri, "/") {
		urlStr += uri
	} else {
		urlStr += "/" + uri
	}

	var body io.Reader
	if jsonBody != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(jsonBody); err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("encoding json: %w", err)
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("preparing http request: %w", err)
	}
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	req.Header.Set(userAgentHeader, userAgentValue)
	c.addCSRFHeader(req)

	client := c.getHTTPClient()

	resp, err := client.Do(req)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("sending http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("reading http response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		c.authMu.Lock()
		c.isAuthenticated = false
		c.cookie = http.Cookie{}
		c.authMu.Unlock()

		if err := c.authenticate(ctx); err != nil {
			return "", http.StatusUnauthorized, fmt.Errorf("re-authentication failed: %w", err)
		}

		// Recreate the body reader for the retry request
		var retryBody io.Reader
		if jsonBody != nil {
			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(jsonBody); err != nil {
				return "", http.StatusInternalServerError, fmt.Errorf("encoding json for retry: %w", err)
			}
			retryBody = buf
		}

		req2, err := http.NewRequestWithContext(ctx, method, urlStr, retryBody)
		if err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("creating http request after reauth: %w", err)
		}
		req2.Header.Set(contentTypeHeader, contentTypeJSON)
		req2.Header.Set(userAgentHeader, userAgentValue)
		c.addCSRFHeader(req2)

		client2 := c.getHTTPClient()

		resp2, err := client2.Do(req2)
		if err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("sending http request after reauth: %w", err)
		}

		defer resp2.Body.Close()

		respBody, err = io.ReadAll(resp2.Body)
		if err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("reading http response after reauth: %w", err)
		}

		return string(respBody), resp2.StatusCode, nil
	}

	// Handle 403 Forbidden - CSRF token invalid/expired, refresh token and retry
	if resp.StatusCode == http.StatusForbidden && c.csrfEnabled && c.csrfToken != "" {
		// Clear CSRF token and extract new one from response
		c.csrfToken = ""
		c.extractCSRFToken(resp) // Try to extract new token

		// If we still have no CSRF token after extraction, retry with what we have
		// If we do have a new CSRF token, add it to the request
		c.addCSRFHeader(req)

		// Retry the original request with the new CSRF token
		resp2, err := c.getHTTPClient().Do(req)
		if err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("retrying request after CSRF refresh: %w", err)
		}
		defer resp2.Body.Close()

		respBody, err = io.ReadAll(resp2.Body)
		if err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("reading http response after CSRF refresh: %w", err)
		}

		return string(respBody), resp2.StatusCode, nil
	}

	return string(respBody), resp.StatusCode, nil
}
