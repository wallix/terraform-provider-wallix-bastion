package bastion

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sync"

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

// sharedClients holds one client per provider configuration, keyed by its hash so that no secret
// is kept as a key: the SDKv2 and Framework servers are configured separately.
var sharedClients = struct { //nolint:gochecknoglobals
	sync.Mutex

	byConfig map[[sha256.Size]byte]*Client
}{byConfig: make(map[[sha256.Size]byte]*Client)}

// sharedClient checks config and returns its client, created on first use, so that both servers
// share one Bastion session and one connection pool.
func sharedClient(config Config) (*Client, diag.Diagnostics) {
	if config.BastionIP == "" {
		return nil, diag.Errorf("missing 'ip' configuration to configure provider")
	}
	if config.BastionUser == "" {
		return nil, diag.Errorf("missing 'user' configuration to configure provider")
	}
	if config.BastionPort < 0 || config.BastionPort > math.MaxUint16 {
		return nil, diag.Errorf("invalid value %d for 'port' configuration to configure provider", config.BastionPort)
	}

	key := sha256.Sum256(fmt.Appendf(nil, "%#v", config))
	sharedClients.Lock()
	defer sharedClients.Unlock()
	if client, ok := sharedClients.byConfig[key]; ok {
		return client, nil
	}
	client, diags := config.Client()
	if diags.HasError() {
		return nil, diags
	}
	sharedClients.byConfig[key] = client

	return client, diags
}
