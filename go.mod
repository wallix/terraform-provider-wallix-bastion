module github.com/claranet/terraform-provider-wallix-bastion

go 1.16

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.8.0
	github.com/jeremmfr/go-utils v0.3.0
)

replace github.com/hashicorp/terraform-plugin-sdk/v2 v2.8.0 => github.com/jeremmfr/terraform-plugin-sdk/v2 v2.8.1-0.20211007115003-2ac7d96a040a
