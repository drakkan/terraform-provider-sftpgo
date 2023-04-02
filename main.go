package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name sftpgo

func main() {
	providerserver.Serve(context.Background(), sftpgo.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/drakkan/sftpgo",
	})
}
