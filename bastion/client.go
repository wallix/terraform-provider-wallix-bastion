package bastion

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
)

// Information to connect on Wallix bastion.
type Client struct {
	bastionPort       int
	bastionAPIVersion string
	bastionIP         string
	bastionToken      string
	bastionUser       string
	bastionPwd        string
}

var defaultHTTPClient *http.Client //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint: gosec
	defaultHTTPClient = &http.Client{Transport: transport}
}

func (c *Client) newRequest(ctx context.Context, uri string, method string, jsonBody interface{}) (string, int, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(jsonBody)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("decoding json: %w", err)
	}
	url := "https://" + c.bastionIP + ":" + strconv.Itoa(c.bastionPort) + "/api/" + c.bastionAPIVersion
	if strings.HasPrefix(uri, "/") {
		url += uri
	} else {
		url += "/" + uri
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "terraform-provider-wallix-bastion")
	if c.bastionToken != "" {
		req.Header.Add("X-Auth-Key", c.bastionToken)
		req.Header.Add("X-Auth-User", c.bastionUser)
	} else {
		rawcreds := c.bastionUser + ":" + c.bastionPwd
		encodedcreds := base64.StdEncoding.EncodeToString([]byte(rawcreds))
		req.Header.Add("Authorization", "Basic "+encodedcreds)
	}
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("preparing http request: %w", err)
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("sending http request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("reading http response: %w", err)
	}

	return string(respBody), resp.StatusCode, nil
}
