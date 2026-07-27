package bastion

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Config: provider config.
type Config struct {
	BastionPort        int
	BastionAPIVersion  string
	BastionIP          string
	BastionToken       string
	BastionUser        string
	BastionPwd         string
	SessionTimeout     int
	CSRFEnabled        bool
	InsecureSkipVerify bool
}

// Client: read information to connect on wallix bastion.
func (c *Config) Client() (*Client, diag.Diagnostics) {
	cl := &Client{
		bastionIP:          c.BastionIP,
		bastionPort:        c.BastionPort,
		bastionToken:       c.BastionToken,
		bastionUser:        c.BastionUser,
		bastionAPIVersion:  c.BastionAPIVersion,
		bastionPwd:         c.BastionPwd,
		sessionTimeout:     c.SessionTimeout,
		csrfEnabled:        c.CSRFEnabled,
		insecureSkipVerify: c.InsecureSkipVerify,
	}

	return cl, nil
}
