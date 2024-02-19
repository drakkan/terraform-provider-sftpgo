package sftpgo

import (
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// we have to set the following env vars to run tests
// SFTPGO_HOST="http://<ip>:8080"
// SFTPGO_USERNAME=...
// SFTPGO_PASSWORD=...

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sftpgo": providerserver.NewProtocol6WithError(New()),
	}
)

func getClient() (*client.Client, error) {
	host := os.Getenv("SFTPGO_HOST")
	user := os.Getenv("SFTPGO_USERNAME")
	pwd := os.Getenv("SFTPGO_PASSWORD")

	headers := map[string]string{
		"FOO": "BAR",
	}
	return client.NewClient(&host, &user, &pwd, nil, headers)
}
