package bastion

import (
	"testing"
)

const testHost = "bastion.wallix.local"

func TestTLSInsecureSkipVerifyDefault(t *testing.T) {
	config := &Config{
		BastionIP:          testHost,
		BastionPort:        443,
		BastionUser:        "admin",
		BastionPwd:         skPassword,
		InsecureSkipVerify: false, // Default: secure
	}

	client, _ := config.Client()
	if client.insecureSkipVerify != false {
		t.Errorf("Expected insecureSkipVerify to be false (secure by default), got %v", client.insecureSkipVerify)
	}
}

func TestTLSInsecureSkipVerifyEnabled(t *testing.T) {
	config := &Config{
		BastionIP:          testHost,
		BastionPort:        443,
		BastionUser:        "admin",
		BastionPwd:         skPassword,
		InsecureSkipVerify: true, // Allow self-signed for dev
	}

	client, _ := config.Client()
	if client.insecureSkipVerify != true {
		t.Errorf("Expected insecureSkipVerify to be true, got %v", client.insecureSkipVerify)
	}
}

func TestTLSHTTPTransportConfiguration(t *testing.T) {
	config := &Config{
		BastionIP:          testHost,
		BastionPort:        443,
		BastionUser:        "admin",
		BastionPwd:         skPassword,
		InsecureSkipVerify: false,
	}

	client, _ := config.Client()
	client.initJar()

	httpClient := client.getHTTPClient()
	if httpClient == nil {
		t.Fatal("Expected HTTP client to be non-nil")
	}

	if httpClient.Jar == nil {
		t.Fatal("Expected HTTP client jar to be non-nil")
	}

	if httpClient.Transport == nil {
		t.Fatal("Expected HTTP client transport to be non-nil")
	}
}
