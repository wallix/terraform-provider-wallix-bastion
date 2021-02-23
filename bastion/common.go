package bastion

type jsonRestriction struct {
	Action      string `json:"action"`
	Rules       string `json:"rules"`
	SubProtocol string `json:"subprotocol"`
}
