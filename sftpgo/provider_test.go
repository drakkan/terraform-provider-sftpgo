package sftpgo

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

func TestGetEnvAsBool(t *testing.T) {
	os.Unsetenv("TEST_BOOL_ENV")

	tests := []struct {
		name        string
		envValue    string
		defaultVal  bool
		expected    bool
	}{
		{
			name:       "default value when unset",
			envValue:   "",
			defaultVal: true,
			expected:   true,
		},
		{
			name:       "default value when empty",
			envValue:   "",
			defaultVal: false,
			expected:   false,
		},
		{
			name:       "true from environment",
			envValue:   "true",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "false from environment",
			envValue:   "false",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "1 as true",
			envValue:   "1",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "0 as false",
			envValue:   "0",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "invalid value returns default",
			envValue:   "invalid",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL_ENV", tt.envValue)
			} else {
				os.Unsetenv("TEST_BOOL_ENV")
			}

			result := getEnvAsBool("TEST_BOOL_ENV", tt.defaultVal)
			require.Equal(t, tt.expected, result)
		})
	}
}

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
	headers := getHeadersFromEnv()
	edition := getIntFromEnv("SFTPGO_EDITION", 0)
	tlsVerification := getEnvAsBool("SFTPGO_TLS_VERIFICATION", true)

	return client.NewClient(host, user, pwd, "", headers, edition, tlsVerification)
}
