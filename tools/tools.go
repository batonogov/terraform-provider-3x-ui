//go:build tools

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-framework/providerserver"
	_ "github.com/hashicorp/terraform-plugin-go/tfprotov6"
	_ "github.com/hashicorp/terraform-plugin-log/tflog"
)
