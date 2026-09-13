// Copyright (C) 2025 Nicola Murino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetAccessToken(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		expiresAt time.Time
		want      string
	}{
		{name: "missing expiry"},
		{name: "expired long ago", expiresAt: now.Add(-3 * time.Minute)},
		{name: "recently expired", expiresAt: now.Add(-time.Minute)},
		{name: "expires now", expiresAt: now},
		{name: "expires within refresh window", expiresAt: now.Add(time.Minute)},
		{name: "at refresh boundary", expiresAt: now.Add(2 * time.Minute)},
		{name: "valid beyond refresh window", expiresAt: now.Add(5 * time.Minute), want: "cached-token"},
	}

	t.Run("no cached response", func(t *testing.T) {
		c := &Client{}
		require.Empty(t, c.getAccessToken())
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			c.setAuthResponse(&AuthResponse{AccessToken: "cached-token", ExpiresAt: tt.expiresAt})
			require.Equal(t, tt.want, c.getAccessToken())
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "StatusError with 404",
			err:      &StatusError{StatusCode: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "StatusError with 500",
			err:      &StatusError{StatusCode: http.StatusInternalServerError},
			expected: false,
		},
		{
			name:     "Non-StatusError",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "StatusError as value (not pointer)",
			err:      StatusError{StatusCode: http.StatusNotFound},
			expected: false, // IsNotFound only matches *StatusError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

func TestNewClientTLSVerification(t *testing.T) {
	tests := []struct {
		name           string
		tlsSkipVerify  bool
		expectInsecure bool
	}{
		{
			name:           "TLS verification enabled",
			tlsSkipVerify:  false,
			expectInsecure: false,
		},
		{
			name:           "TLS verification disabled",
			tlsSkipVerify:  true,
			expectInsecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient("http://localhost:8080", "admin", "password", "", nil, 0, tt.tlsSkipVerify)
			require.NoError(t, err)
			require.NotNil(t, client)

			if tt.expectInsecure {
				require.NotNil(t, client.HTTPClient.Transport)
				transport, ok := client.HTTPClient.Transport.(*http.Transport)
				require.True(t, ok, "Expected *http.Transport")
				require.NotNil(t, transport.TLSClientConfig)
				require.True(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
			} else {
				if client.HTTPClient.Transport != nil {
					transport, ok := client.HTTPClient.Transport.(*http.Transport)
					if ok && transport.TLSClientConfig != nil {
						require.False(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be false")
					}
				}
			}
		})
	}
}
