package bastion

type jsonApplicationLocalDomain struct {
	EnablePasswordChange           bool                    `json:"enable_password_change"`
	ID                             string                  `json:"id,omitempty"`
	DomainName                     string                  `json:"domain_name"`
	Description                    string                  `json:"description"`
	PasswordChangePolicy           string                  `json:"password_change_policy,omitempty"`
	PasswordChangePlugin           string                  `json:"password_change_plugin,omitempty"`
	PasswordChangePluginParameters *map[string]interface{} `json:"password_change_plugin_parameters,omitempty"`
}
