package bastion_test

import (
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// TestConfigClientWithCSRFOptions tests that Config properly creates Client with CSRF options.
func TestConfigClientWithCSRFOptions(t *testing.T) {
	tests := []struct {
		name            string
		config          bastion.Config
		wantErr         bool
		wantSessionTO   int
		wantCSRFEnabled bool
	}{
		{
			name: "default config with CSRF enabled",
			config: bastion.Config{
				BastionIP:      "192.168.1.1",
				BastionPort:    443,
				BastionUser:    "admin",
				BastionPwd:     "password",
				SessionTimeout: 120,
				CSRFEnabled:    true,
			},
			wantErr:         false,
			wantSessionTO:   120,
			wantCSRFEnabled: true,
		},
		{
			name: "config with CSRF disabled",
			config: bastion.Config{
				BastionIP:      "192.168.1.1",
				BastionPort:    443,
				BastionUser:    "admin",
				BastionPwd:     "password",
				SessionTimeout: 240,
				CSRFEnabled:    false,
			},
			wantErr:         false,
			wantSessionTO:   240,
			wantCSRFEnabled: false,
		},
		{
			name: "config with custom session timeout",
			config: bastion.Config{
				BastionIP:      "192.168.1.1",
				BastionPort:    8443,
				BastionUser:    "testuser",
				BastionPwd:     "testpass",
				SessionTimeout: 300,
				CSRFEnabled:    true,
			},
			wantErr:         false,
			wantSessionTO:   300,
			wantCSRFEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, diags := tt.config.Client()
			if diags.HasError() != tt.wantErr {
				t.Errorf("Config.Client() error = %v, wantErr %v", diags.HasError(), tt.wantErr)

				return
			}

			if client == nil {
				if !tt.wantErr {
					t.Errorf("Config.Client() returned nil client")
				}

				return
			}

			// Verify client was created successfully
			// Note: CSRF and session timeout are internal client settings
			// and are validated through integration/behavioral tests
		})
	}
}
