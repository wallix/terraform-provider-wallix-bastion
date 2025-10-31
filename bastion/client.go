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

type Client struct {
	bastionPort       int
	bastionAPIVersion string
	bastionIP         string
	bastionToken      string
	bastionUser       string
	bastionPwd        string

	jar             http.CookieJar
	cookie          http.Cookie
	isAuthenticated bool
	authMu          sync.Mutex
}

var defaultHTTPClient *http.Client //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	defaultHTTPClient = &http.Client{Transport: transport}
}

func (c *Client) initJar() {
	if c.jar == nil {
		c.jar, _ = cookiejar.New(nil)
	}
}

func (c *Client) cookieValid() bool {
	if c.cookie.Name != "wab_session_id" || c.cookie.Value == "" {
		return false
	}

	return true
}

func (c *Client) authenticate() error {
	c.authMu.Lock()

	defer c.authMu.Unlock()

	if c.isAuthenticated && c.cookieValid() {
		return nil
	}

	c.initJar()

	authURL := fmt.Sprintf("https://%s/api", net.JoinHostPort(c.bastionIP, strconv.Itoa(c.bastionPort)))
	client := *defaultHTTPClient
	client.Jar = c.jar
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, nil)
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-provider-wallix-bastion")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

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

	u, _ := url.Parse(authURL)
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
	c.isAuthenticated = true

	return nil
}

func (c *Client) newRequest(ctx context.Context, uri, method string, jsonBody interface{}) (string, int, error) {
	if !c.isAuthenticated || !c.cookieValid() {
		if err := c.authenticate(); err != nil {
			return "", http.StatusUnauthorized, fmt.Errorf("authentication failed: %w", err)
		}
	}

	c.initJar()

	urlStr := fmt.Sprintf("https://%s/api/%s", net.JoinHostPort(c.bastionIP, strconv.Itoa(c.bastionPort)), c.bastionAPIVersion)
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-wallix-bastion")

	client := *defaultHTTPClient
	client.Jar = c.jar

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

		if err := c.authenticate(); err != nil {
			return "", http.StatusUnauthorized, fmt.Errorf("re-authentication failed: %w", err)
		}

		req2, _ := http.NewRequestWithContext(ctx, method, urlStr, body)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("User-Agent", "terraform-provider-wallix-bastion")

		client2 := *defaultHTTPClient
		client2.Jar = c.jar

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

	return string(respBody), resp.StatusCode, nil
}
