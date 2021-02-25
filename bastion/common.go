package bastion

type jsonRestriction struct {
	Action      string `json:"action"`
	Rules       string `json:"rules"`
	SubProtocol string `json:"subprotocol"`
}

type jsonCredential struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}
