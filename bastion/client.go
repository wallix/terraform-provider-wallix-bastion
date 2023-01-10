package bastion

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Information to connect on Wallix bastion.
type Client struct {
	bastionPort       int
	bastionAPIVersion string
	bastionIP         string
	bastionToken      string
	bastionUser       string
}

func (c *Client) newRequest(ctx context.Context, uri string, method string, jsonBody interface{}) (string, int, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(jsonBody)
	if err != nil {
		return "", http.StatusInternalServerError, err
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
	req.Header.Add("X-Auth-Key", c.bastionToken)
	req.Header.Add("X-Auth-User", c.bastionUser)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return string(respBody), resp.StatusCode, nil
}
