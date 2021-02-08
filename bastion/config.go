package bastion

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Config : provider config.
type Config struct {
	bastionPort    int
	bastionIP      string
	bastionToken   string
	bastionUser    string
	bastionVersion string
}

// Client : read information to connect on wallix bastion.
func (c *Config) Client() (*Client, diag.Diagnostics) {
	cl := &Client{
		bastionIP:      c.bastionIP,
		bastionPort:    c.bastionPort,
		bastionToken:   c.bastionToken,
		bastionUser:    c.bastionUser,
		bastionVersion: c.bastionVersion,
	}

	return cl, nil
}
